/**
 * Program IDL in camelCase format in order to be used in JS/TS.
 *
 * Note that this is only a type helper and is not the actual IDL. The original
 * IDL can be found at `target/idl/transfer.json`.
 */
export type Transfer = {
  "address": "Fa5VLqopKp6cokXJreYeNNmUG8F9AaE4CUBnGQvtdq7Q",
  "metadata": {
    "name": "transfer",
    "version": "0.1.0",
    "spec": "0.1.0",
    "description": "Created with Anchor"
  },
  "instructions": [
    {
      "name": "batchTransfer",
      "discriminator": [
        209,
        90,
        4,
        108,
        61,
        185,
        18,
        139
      ],
      "accounts": [
        {
          "name": "sourceOwner",
          "docs": [
            "0 — Token holder authorising the batch.",
            "Token-2022's `transfer_checked` validates that this matches `source.owner`."
          ],
          "signer": true
        },
        {
          "name": "source",
          "docs": [
            "1 — Source token account (shared by every leg)."
          ],
          "writable": true
        },
        {
          "name": "mint",
          "docs": [
            "2 — The Token-2022 mint."
          ]
        },
        {
          "name": "extraAccountMetaList",
          "pda": {
            "seeds": [
              {
                "kind": "const",
                "value": [
                  101,
                  120,
                  116,
                  114,
                  97,
                  45,
                  97,
                  99,
                  99,
                  111,
                  117,
                  110,
                  116,
                  45,
                  109,
                  101,
                  116,
                  97,
                  115
                ]
              },
              {
                "kind": "account",
                "path": "mint"
              }
            ],
            "program": {
              "kind": "const",
              "value": [
                27,
                88,
                12,
                56,
                186,
                148,
                80,
                211,
                123,
                61,
                151,
                212,
                88,
                255,
                3,
                60,
                148,
                129,
                56,
                0,
                170,
                239,
                39,
                70,
                25,
                205,
                76,
                74,
                227,
                178,
                10,
                177
              ]
            }
          }
        },
        {
          "name": "transferHookProgram",
          "address": "2qjsucJfrjP93FCwnYjc9EjYzYS8u31eWHhQo1jR9pcg"
        },
        {
          "name": "freezeProgram",
          "address": "8L1kqDvAYC9dQXNNNnZbABtRbHGjzoxSgAPzbQZmwmSd"
        },
        {
          "name": "deployProgram",
          "address": "HCe5Um7ThFBzDSyn256EPQvyr6jy6E66ydzZ5hMta3Tq"
        },
        {
          "name": "assetConfigurationPda",
          "pda": {
            "seeds": [
              {
                "kind": "const",
                "value": [
                  97,
                  115,
                  115,
                  101,
                  116,
                  95,
                  99,
                  111,
                  110,
                  102,
                  105,
                  103,
                  117,
                  114,
                  97,
                  116,
                  105,
                  111,
                  110
                ]
              },
              {
                "kind": "account",
                "path": "mint"
              }
            ],
            "program": {
              "kind": "const",
              "value": [
                240,
                182,
                77,
                149,
                174,
                242,
                208,
                54,
                31,
                138,
                200,
                30,
                243,
                31,
                148,
                113,
                161,
                240,
                63,
                40,
                108,
                82,
                42,
                48,
                110,
                115,
                30,
                160,
                98,
                125,
                248,
                52
              ]
            }
          }
        },
        {
          "name": "factoryProgram",
          "address": "FEY9E77nH7R1gLGNxkhYKchJpB6MgpMrWMhkNXrNhzR5"
        },
        {
          "name": "assetClassVersionPda",
          "pda": {
            "seeds": [
              {
                "kind": "const",
                "value": [
                  97,
                  115,
                  115,
                  101,
                  116,
                  95,
                  99,
                  108,
                  97,
                  115,
                  115,
                  95,
                  118,
                  101,
                  114,
                  115,
                  105,
                  111,
                  110
                ]
              },
              {
                "kind": "account",
                "path": "asset_configuration_pda.asset_class_config_id",
                "account": "assetConfiguration"
              },
              {
                "kind": "account",
                "path": "asset_configuration_pda.asset_class_version_id",
                "account": "assetConfiguration"
              }
            ],
            "program": {
              "kind": "const",
              "value": [
                211,
                123,
                97,
                49,
                148,
                61,
                0,
                123,
                122,
                206,
                49,
                38,
                235,
                135,
                128,
                117,
                92,
                177,
                90,
                249,
                180,
                1,
                59,
                22,
                82,
                69,
                14,
                21,
                75,
                204,
                57,
                168
              ]
            }
          }
        },
        {
          "name": "deactivateProgram",
          "address": "H2iRjVVKsKQMAnJKqiTfW2LGvT1G9tDqQ81DzRjxfX7V"
        },
        {
          "name": "deactivatePda",
          "pda": {
            "seeds": [
              {
                "kind": "const",
                "value": [
                  100,
                  101,
                  97,
                  99,
                  116,
                  105,
                  118,
                  97,
                  116,
                  101
                ]
              },
              {
                "kind": "account",
                "path": "mint"
              }
            ],
            "program": {
              "kind": "const",
              "value": [
                238,
                43,
                105,
                162,
                163,
                237,
                139,
                95,
                81,
                45,
                69,
                187,
                195,
                53,
                53,
                92,
                147,
                252,
                127,
                84,
                5,
                224,
                26,
                131,
                233,
                21,
                47,
                51,
                149,
                90,
                20,
                208
              ]
            }
          }
        },
        {
          "name": "transferControlProgram",
          "address": "3h92PdZJB7TuCzp6iPDtrJm2k8V7fn5ETYNwCYiYy9Eo"
        },
        {
          "name": "transferControlModePda",
          "pda": {
            "seeds": [
              {
                "kind": "const",
                "value": [
                  116,
                  114,
                  97,
                  110,
                  115,
                  102,
                  101,
                  114,
                  95,
                  99,
                  111,
                  110,
                  116,
                  114,
                  111,
                  108,
                  95,
                  109,
                  111,
                  100,
                  101
                ]
              },
              {
                "kind": "account",
                "path": "mint"
              }
            ],
            "program": {
              "kind": "const",
              "value": [
                39,
                255,
                173,
                207,
                102,
                3,
                77,
                127,
                203,
                150,
                188,
                14,
                86,
                207,
                133,
                158,
                217,
                57,
                254,
                20,
                232,
                146,
                113,
                164,
                190,
                234,
                43,
                128,
                246,
                189,
                25,
                208
              ]
            }
          }
        },
        {
          "name": "sourceWhitelistPda",
          "pda": {
            "seeds": [
              {
                "kind": "const",
                "value": [
                  119,
                  104,
                  105,
                  116,
                  101,
                  108,
                  105,
                  115,
                  116
                ]
              },
              {
                "kind": "account",
                "path": "mint"
              },
              {
                "kind": "account",
                "path": "source"
              }
            ],
            "program": {
              "kind": "const",
              "value": [
                39,
                255,
                173,
                207,
                102,
                3,
                77,
                127,
                203,
                150,
                188,
                14,
                86,
                207,
                133,
                158,
                217,
                57,
                254,
                20,
                232,
                146,
                113,
                164,
                190,
                234,
                43,
                128,
                246,
                189,
                25,
                208
              ]
            }
          }
        },
        {
          "name": "sourceFrozenPda",
          "pda": {
            "seeds": [
              {
                "kind": "const",
                "value": [
                  102,
                  114,
                  111,
                  122,
                  101,
                  110,
                  95,
                  97,
                  99,
                  99,
                  111,
                  117,
                  110,
                  116
                ]
              },
              {
                "kind": "account",
                "path": "mint"
              },
              {
                "kind": "account",
                "path": "source"
              }
            ],
            "program": {
              "kind": "const",
              "value": [
                108,
                224,
                178,
                198,
                200,
                195,
                222,
                160,
                194,
                118,
                224,
                225,
                141,
                84,
                199,
                6,
                16,
                25,
                118,
                188,
                147,
                191,
                37,
                168,
                240,
                56,
                239,
                183,
                19,
                111,
                180,
                238
              ]
            }
          }
        },
        {
          "name": "sourceFrozenBalancePda",
          "pda": {
            "seeds": [
              {
                "kind": "const",
                "value": [
                  102,
                  114,
                  111,
                  122,
                  101,
                  110,
                  95,
                  98,
                  97,
                  108,
                  97,
                  110,
                  99,
                  101
                ]
              },
              {
                "kind": "account",
                "path": "mint"
              },
              {
                "kind": "account",
                "path": "source"
              }
            ],
            "program": {
              "kind": "const",
              "value": [
                108,
                224,
                178,
                198,
                200,
                195,
                222,
                160,
                194,
                118,
                224,
                225,
                141,
                84,
                199,
                6,
                16,
                25,
                118,
                188,
                147,
                191,
                37,
                168,
                240,
                56,
                239,
                183,
                19,
                111,
                180,
                238
              ]
            }
          }
        },
        {
          "name": "holdProgram",
          "address": "J8iq5Qz8tXLswZBbUFHuJukf3jpwEXLGVpvFoPZb2qY3"
        },
        {
          "name": "sourceHoldPositionPda",
          "pda": {
            "seeds": [
              {
                "kind": "const",
                "value": [
                  104,
                  111,
                  108,
                  100,
                  95,
                  112,
                  111,
                  115,
                  105,
                  116,
                  105,
                  111,
                  110
                ]
              },
              {
                "kind": "account",
                "path": "mint"
              },
              {
                "kind": "account",
                "path": "source"
              }
            ],
            "program": {
              "kind": "const",
              "value": [
                254,
                145,
                10,
                28,
                129,
                235,
                144,
                157,
                4,
                171,
                210,
                253,
                129,
                1,
                70,
                47,
                125,
                120,
                181,
                218,
                92,
                202,
                145,
                28,
                131,
                177,
                28,
                183,
                244,
                203,
                112,
                144
              ]
            }
          }
        },
        {
          "name": "token2022Program",
          "address": "TokenzQdBNbLqP5VEhdkAS6EPFLC1PHnBqCXEpPxuEb"
        }
      ],
      "args": [
        {
          "name": "amounts",
          "type": {
            "vec": "u64"
          }
        }
      ]
    }
  ],
  "errors": [
    {
      "code": 6000,
      "name": "emptyBatch",
      "msg": "The batch must contain at least one destination"
    },
    {
      "code": 6001,
      "name": "invalidRemainingAccounts",
      "msg": "Expected exactly one remaining account (destination) per amount for transfer, or two (destination + whitelist PDA) per amount for verify"
    },
    {
      "code": 6002,
      "name": "batchAmountOverflow",
      "msg": "The sum of the batch amounts overflows u64"
    }
  ],
  "types": [
    {
      "name": "assetConfiguration",
      "type": {
        "kind": "struct",
        "fields": [
          {
            "name": "assetClassConfigId",
            "type": "u64"
          },
          {
            "name": "assetClassVersionId",
            "type": "u64"
          },
          {
            "name": "bump",
            "type": "u8"
          }
        ]
      }
    }
  ]
};
