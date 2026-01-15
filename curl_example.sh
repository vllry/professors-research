#!/bin/bash

# Example curl command for /prize-odds endpoint with CardSets
# This demonstrates a simple decklist with CardSet combinations

curl -X POST http://localhost:8080/prize-odds \
  -H "Content-Type: application/json" \
  -d '{
    "decklist": "Pokémon: 12\n1 Bloodmoon Ursaluna ex PRE 168\n1 Hawlucha SVI 118\n4 Drakloak TWM 129\n4 Dreepy TWM 128\n1 Munkidori TWM 95\n3 Dragapult ex TWM 130\n1 Latias ex SSP 76\n2 Dusclops PRE 36\n2 Duskull SFA 18\n1 Fezandipiti ex SFA 38\n2 Budew PRE 4\n1 Dusknoir SFA 20\n\nTrainer: 11\n4 Ultra Ball SVI 196\n4 Buddy-Buddy Poffin TEF 144\n4 Lillie'\''s Determination MEG 119\n3 Counter Catcher PAR 160\n2 Jamming Tower TWM 153\n4 Iono PAL 185\n2 Night Stretcher SFA 61\n1 Professor Turo'\''s Scenario PRE 121\n1 Nest Ball SVI 181\n3 Boss'\''s Orders PAL 172\n2 Hilda WHT 84\n\nEnergy: 4\n1 Basic {R} Energy EVO 92\n2 Basic {P} Energy EVO 95\n3 Luminous Energy PAL 191\n1 Neo Upper Energy TEF 162\n\nTotal Cards: 60",
    "cardSets": {
      "main_line_all_prized": [
        {
          "anyOfs": [
            {
              "cards": [
                {
                  "card": {
                    "name": "Dragapult ex",
                    "setCode": "TWM",
                    "number": "130"
                  },
                  "count": 3
                },
                {
                  "card": {
                    "name": "Drakloak",
                    "setCode": "TWM",
                    "number": "129"
                  },
                  "count": 4
                },
                {
                  "card": {
                    "name": "Dreepy",
                    "setCode": "TWM",
                    "number": "128"
                  },
                  "count": 4
                }
              ]
            }
          ]
        }
      ],
      "limited_dragapult": [
        {
          "anyOfs": [
            {
              "cards": [
                {
                  "card": {
                    "name": "Dragapult ex",
                    "setCode": "TWM",
                    "number": "130"
                  },
                  "count": 2
                }
              ]
            }
          ]
        },
        {
          "allOfs": [
            {
              "cards": [
                {
                  "card": {
                    "name": "Night Stretcher",
                    "setCode": "SFA",
                    "number": "61"
                  },
                  "count": 2
                },
                {
                  "card": {
                    "name": "Dragapult ex",
                    "setCode": "TWM",
                    "number": "130"
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
