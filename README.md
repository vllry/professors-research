This repo contains multiple tools to help with Pokemon TCG optimization.

## API Server

The API server provides a REST API for calculating prize odds from Pokemon TCG decklists.

### Building and Running

Build the server:
```bash
make build
```

Run the server:
```bash
make run
```

Or set a custom port:
```bash
PORT=3000 make run
```

### API Endpoints

#### Health Check

```bash
curl http://localhost:8080/api/health
```

#### Calculate Prize Odds

Calculate prize odds for a decklist:

```bash
curl -X POST http://localhost:8080/api/prize-odds \
  -H "Content-Type: application/json" \
  -d '{
    "decklist": "Pokémon: 12\n1 Bloodmoon Ursaluna ex PRE 168\n1 Hawlucha SVI 118\n4 Drakloak TWM 129\n4 Dreepy TWM 128\n1 Munkidori TWM 95\n3 Dragapult ex TWM 130\n1 Latias ex SSP 76\n2 Dusclops PRE 36\n2 Duskull SFA 18\n1 Fezandipiti ex SFA 38\n2 Budew PRE 4\n1 Dusknoir SFA 20\n\nTrainer: 11\n4 Ultra Ball SVI 196\n4 Buddy-Buddy Poffin TEF 144\n4 Lillie'\''s Determination MEG 119\n3 Counter Catcher PAR 160\n2 Jamming Tower TWM 153\n4 Iono PAL 185\n2 Night Stretcher SFA 61\n1 Professor Turo'\''s Scenario PRE 121\n1 Nest Ball SVI 181\n3 Boss'\''s Orders PAL 172\n2 Hilda WHT 84\n\nEnergy: 4\n1 Basic {R} Energy EVO 92\n2 Basic {P} Energy EVO 95\n3 Luminous Energy PAL 191\n1 Neo Upper Energy TEF 162\n\nTotal Cards: 60"
  }'
```

Or using a file with `jq`:

```bash
jq -Rs '{decklist: .}' testdata/dragapult_MEG_1.txt | \
  curl -X POST http://localhost:8080/api/prize-odds \
    -H "Content-Type: application/json" \
    -d @-
```

The response contains a map of card names to arrays of cumulative odds. Each array contains the probability of prizing **at least** X copies:
- Index 0: probability of prizing at least 1 copy
- Index 1: probability of prizing at least 2 copies
- Index 2: probability of prizing at least 3 copies
- ... and so on, up to min(number of copies in deck, 6)

For example, a card with 1 copy returns `[0.1]` meaning:
- 10% chance of prizing at least 1 copy (i.e., the single copy)

A card with 4 copies returns `[0.351, 0.046, 0.002, 0.00003]` meaning:
- 35.1% chance of prizing at least 1 copy
- 4.6% chance of prizing at least 2 copies
- 0.2% chance of prizing at least 3 copies
- 0.003% chance of prizing at least 4 copies

```json
{
  "odds": {
    "Bloodmoon Ursaluna ex PRE 168": [0.1],
    "Basic {P} Energy EVO 95": [0.192, 0.008],
    ...
  }
}
```

## Download Tournament (RK9)

Download available **decklists** and **match results** for an RK9 tournament ID:

```bash
go run ./cmd/download-tournament -tournament LV01YShqrqjMo62PxZPg -out ./data/tournaments/LV01YShqrqjMo62PxZPg
```

This writes:
- `tournament.json`: metadata (ID, name, timestamp)
- `decklists.json` + `decklists/*.txt`: parsed decklists (when present on the roster page)
- `matches.json`: parsed match results (when present on the pairings page)

## License

Apache-2.0 (see `LICENSE`).