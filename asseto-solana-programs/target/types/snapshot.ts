/**
 * Program IDL in camelCase format in order to be used in JS/TS.
 *
 * Note that this is only a type helper and is not the actual IDL. The original
 * IDL can be found at `target/idl/snapshot.json`.
 */
export type Snapshot = {
  "address": "hgUtrpstViwxutrkoVXwQh3GQC18wHAmuAvYFTNiV2M",
  "metadata": {
    "name": "snapshot",
    "version": "0.1.0",
    "spec": "0.1.0",
    "description": "Created with Anchor"
  },
  "instructions": [
    {
      "name": "takeSnapshot",
      "discriminator": [
        183,
        210,
        251,
        68,
        140,
        132,
        191,
        140
      ],
      "accounts": [
        {
          "name": "callingAuthority",
          "signer": true
        },
        {
          "name": "payer",
          "writable": true,
          "signer": true
        },
        {
          "name": "mint"
        },
        {
          "name": "snapshotCounter",
          "writable": true,
          "pda": {
            "seeds": [
              {
                "kind": "const",
                "value": [
                  115,
                  110,
                  97,
                  112,
                  115,
                  104,
                  111,
                  116,
                  95,
                  99,
                  111,
                  117,
                  110,
                  116,
                  101,
                  114
                ]
              },
              {
                "kind": "account",
                "path": "mint"
              }
            ]
          }
        },
        {
          "name": "snapshotMerkleRoot",
          "docs": [
            "Immutable Merkle-root PDA for the snapshot being taken.",
            "Seeds: `[\"snapshot_merkle_root\", mint, snapshot_id]`, where `snapshot_id`",
            "is `snapshot_counter.count` (the next-id value read here at account",
            "resolution)."
          ],
          "writable": true,
          "pda": {
            "seeds": [
              {
                "kind": "const",
                "value": [
                  115,
                  110,
                  97,
                  112,
                  115,
                  104,
                  111,
                  116,
                  95,
                  109,
                  101,
                  114,
                  107,
                  108,
                  101,
                  95,
                  114,
                  111,
                  111,
                  116
                ]
              },
              {
                "kind": "account",
                "path": "mint"
              },
              {
                "kind": "account",
                "path": "snapshot_counter.count",
                "account": "snapshotCounter"
              }
            ]
          }
        },
        {
          "name": "systemProgram",
          "address": "11111111111111111111111111111111"
        },
        {
          "name": "eventAuthority"
        },
        {
          "name": "program"
        }
      ],
      "args": [
        {
          "name": "merkleRoot",
          "type": {
            "array": [
              "u8",
              32
            ]
          }
        }
      ]
    }
  ],
  "accounts": [
    {
      "name": "snapshotCounter",
      "discriminator": [
        170,
        211,
        63,
        197,
        196,
        2,
        1,
        56
      ]
    },
    {
      "name": "snapshotMerkleRoot",
      "discriminator": [
        175,
        197,
        152,
        118,
        124,
        199,
        75,
        12
      ]
    }
  ],
  "events": [
    {
      "name": "snapshotTriggered",
      "discriminator": [
        33,
        107,
        94,
        91,
        80,
        61,
        227,
        19
      ]
    }
  ],
  "errors": [
    {
      "code": 6000,
      "name": "unauthorized",
      "msg": "Caller is not an authorised PDA (mint_authority, permanent_delegate, or transfer)"
    },
    {
      "code": 6001,
      "name": "snapshotCounterOverflow",
      "msg": "snapshot counter overflow when creating new snapshot"
    }
  ],
  "types": [
    {
      "name": "snapshotCounter",
      "type": {
        "kind": "struct",
        "fields": [
          {
            "name": "bump",
            "type": "u8"
          },
          {
            "name": "count",
            "docs": [
              "Id of the **next** snapshot (0-based)."
            ],
            "type": "u64"
          }
        ]
      }
    },
    {
      "name": "snapshotMerkleRoot",
      "type": {
        "kind": "struct",
        "fields": [
          {
            "name": "bump",
            "type": "u8"
          },
          {
            "name": "merkleRoot",
            "type": {
              "array": [
                "u8",
                32
              ]
            }
          }
        ]
      }
    },
    {
      "name": "snapshotTriggered",
      "type": {
        "kind": "struct",
        "fields": [
          {
            "name": "mint",
            "type": "pubkey"
          },
          {
            "name": "snapshotId",
            "type": "u64"
          },
          {
            "name": "merkleRoot",
            "type": {
              "array": [
                "u8",
                32
              ]
            }
          }
        ]
      }
    }
  ]
};
