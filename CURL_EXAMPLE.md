# Curl Examples for /prize-odds Endpoint

## Simple Example with CardSets

```bash
curl -X POST http://localhost:8080/prize-odds \
  -H "Content-Type: application/json" \
  -d '{
    "decklist": "Pokémon: 2\n4 Dreepy TWM 128\n4 Drakloak TWM 129\n\nTrainer: 1\n4 Lillies Determination MEG 119\n4 Iono PAL 185\n\nEnergy: 1\n4 Luminous Energy PAL 191\n\nTotal Cards: 60",
    "cardSets": {
      "draw_support": [
        {
          "anyOfs": [
            {
              "cards": [
                {
                  "card": {
                    "name": "Lillies Determination",
                    "setCode": "MEG",
                    "number": "119"
                  },
                  "count": 1
                },
                {
                  "card": {
                    "name": "Iono",
                    "setCode": "PAL",
                    "number": "185"
                  },
                  "count": 1
                }
              ]
            }
          ]
        }
      ],
      "pokemon_setup": [
        {
          "anyOfs": [
            {
              "cards": [
                {
                  "card": {
                    "name": "Dreepy",
                    "setCode": "TWM",
                    "number": "128"
                  },
                  "count": 1
                }
              ]
            },
            {
              "cards": [
                {
                  "card": {
                    "name": "Drakloak",
                    "setCode": "TWM",
                    "number": "129"
                  },
                  "count": 1
                }
              ]
            }
          ]
        }
      ]
    }
  }'
```

## Minimal Example (CardSets only)

```bash
curl -X POST http://localhost:8080/prize-odds \
  -H "Content-Type: application/json" \
  -d '{
    "decklist": "Pokémon: 1\n4 Test Card TEST 1\n\nTotal Cards: 60",
    "cardSets": {
      "simple": [
        {
          "anyOfs": [
            {
              "cards": [
                {
                  "card": {
                    "name": "Test Card",
                    "setCode": "TEST",
                    "number": "1"
                  },
                  "count": 1
                }
              ]
            }
          ]
        }
      ]
    }
  }'
```

## Response Format

The response will include:
- `odds`: Map of card names to arrays of cumulative odds (existing functionality)
- `cardSetOdds`: Map of CardSet group names to maps of CardSet names to odds

Example response:
```json
{
  "odds": {
    "Test Card TEST 1": [0.1, 0.05, ...]
  },
  "cardSetOdds": {
    "simple": {
      "simple[0]": 0.1
    }
  }
}
```




