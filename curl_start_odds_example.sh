#!/bin/bash

# Example curl command for /api/start-odds endpoint
# Uses the decklist from testdata/dragapult_MEG_1.txt

DECKLIST_FILE="testdata/dragapult_MEG_1.txt"

# Read the decklist file
DECKLIST=$(cat "$DECKLIST_FILE")

# Example 1: Simple request without CardSets
echo "=== Example 1: Simple request without CardSets ==="
curl -X POST http://localhost:8080/api/start-odds \
  -H "Content-Type: application/json" \
  -d "$(jq -n --arg decklist "$DECKLIST" '{decklist: $decklist}')" | jq '.'

echo ""
echo "=== Example 2: Request with CardSets ==="

# Example 2: Request with CardSets
# This checks for:
# - draw_support: At least one draw supporter (Lillie's Determination OR Iono) in starting hand
# - basic_starters: At least one basic Pokemon starter (Dreepy OR Duskull OR Budew) in starting hand
curl -X POST http://localhost:8080/api/start-odds \
  -H "Content-Type: application/json" \
  -d "$(jq -n \
    --arg decklist "$DECKLIST" \
    '{
      decklist: $decklist,
      cardSets: {
        draw_support: [
          {
            anyOfs: [
              {
                cards: [
                  {
                    card: {
                      name: "Lillie'\''s Determination",
                      setCode: "MEG",
                      number: "119"
                    },
                    count: 1
                  },
                  {
                    card: {
                      name: "Iono",
                      setCode: "PAL",
                      number: "185"
                    },
                    count: 1
                  }
                ]
              }
            ]
          }
        ],
        basic_starters: [
          {
            anyOfs: [
              {
                cards: [
                  {
                    card: {
                      name: "Dreepy",
                      setCode: "TWM",
                      number: "128"
                    },
                    count: 1
                  }
                ]
              },
              {
                cards: [
                  {
                    card: {
                      name: "Duskull",
                      setCode: "SFA",
                      number: "18"
                    },
                    count: 1
                  }
                ]
              },
              {
                cards: [
                  {
                    card: {
                      name: "Budew",
                      setCode: "PRE",
                      number: "4"
                    },
                    count: 1
                  }
                ]
              }
            ]
          }
        ]
      }
    }')" | jq '.'

