/**
 * Program IDL in camelCase format in order to be used in JS/TS.
 *
 * Note that this is only a type helper and is not the actual IDL. The original
 * IDL can be found at `target/idl/hold.json`.
 */
export type Hold = {
  "address": "J8iq5Qz8tXLswZBbUFHuJukf3jpwEXLGVpvFoPZb2qY3",
  "metadata": {
    "name": "hold",
    "version": "0.1.0",
    "spec": "0.1.0",
    "description": "Created with Anchor"
  },
  "instructions": [
    {
      "name": "controllerCreateHold",
      "discriminator": [
        29,
        237,
        146,
        110,
        105,
        20,
        55,
        3
      ],
      "accounts": [
        {
          "name": "payer",
          "writable": true,
          "signer": true
        },
        {
          "name": "authority",
          "signer": true
        },
        {
          "name": "authorityRolesPda",
          "pda": {
            "seeds": [
              {
                "kind": "const",
                "value": [
                  114,
                  111,
                  108,
                  101,
                  115
                ]
              },
              {
                "kind": "account",
                "path": "mint"
              },
              {
                "kind": "account",
                "path": "authority"
              }
            ],
            "program": {
              "kind": "const",
              "value": [
                235,
                41,
                191,
                132,
                186,
                2,
                100,
                233,
                201,
                22,
                224,
                63,
                181,
                155,
                128,
                170,
                45,
                56,
                111,
                156,
                131,
                234,
                141,
                38,
                46,
                129,
                42,
                229,
                87,
                122,
                173,
                44
              ]
            }
          }
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
          "name": "mint"
        },
        {
          "name": "tokenAccount"
        },
        {
          "name": "tokenAccountFrozenPda",
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
                "path": "tokenAccount"
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
          "name": "tokenAccountFrozenBalancePda",
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
                "path": "tokenAccount"
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
          "name": "holdPosition",
          "writable": true,
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
                "path": "tokenAccount"
              }
            ]
          }
        },
        {
          "name": "holdRecord",
          "writable": true,
          "pda": {
            "seeds": [
              {
                "kind": "const",
                "value": [
                  104,
                  111,
                  108,
                  100
                ]
              },
              {
                "kind": "account",
                "path": "mint"
              },
              {
                "kind": "account",
                "path": "tokenAccount"
              },
              {
                "kind": "arg",
                "path": "holdId"
              }
            ]
          }
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
          "name": "tokenProgram"
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
          "name": "holdId",
          "type": "u64"
        },
        {
          "name": "amount",
          "type": "u64"
        },
        {
          "name": "expiration",
          "type": "i64"
        },
        {
          "name": "escrow",
          "type": "pubkey"
        },
        {
          "name": "destination",
          "type": {
            "option": "pubkey"
          }
        }
      ]
    },
    {
      "name": "createHold",
      "discriminator": [
        5,
        206,
        229,
        132,
        223,
        134,
        166,
        221
      ],
      "accounts": [
        {
          "name": "payer",
          "writable": true,
          "signer": true
        },
        {
          "name": "authority",
          "signer": true
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
          "name": "mint"
        },
        {
          "name": "tokenAccount"
        },
        {
          "name": "tokenAccountFrozenPda",
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
                "path": "tokenAccount"
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
          "name": "tokenAccountFrozenBalancePda",
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
                "path": "tokenAccount"
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
          "name": "holdPosition",
          "writable": true,
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
                "path": "tokenAccount"
              }
            ]
          }
        },
        {
          "name": "holdRecord",
          "writable": true,
          "pda": {
            "seeds": [
              {
                "kind": "const",
                "value": [
                  104,
                  111,
                  108,
                  100
                ]
              },
              {
                "kind": "account",
                "path": "mint"
              },
              {
                "kind": "account",
                "path": "tokenAccount"
              },
              {
                "kind": "arg",
                "path": "holdId"
              }
            ]
          }
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
          "name": "tokenProgram"
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
          "name": "holdId",
          "type": "u64"
        },
        {
          "name": "amount",
          "type": "u64"
        },
        {
          "name": "expiration",
          "type": "i64"
        },
        {
          "name": "escrow",
          "type": "pubkey"
        },
        {
          "name": "destination",
          "type": {
            "option": "pubkey"
          }
        }
      ]
    },
    {
      "name": "executeHold",
      "discriminator": [
        97,
        38,
        45,
        206,
        238,
        96,
        237,
        164
      ],
      "accounts": [
        {
          "name": "escrow",
          "signer": true
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
          "name": "mint"
        },
        {
          "name": "sourceToken",
          "writable": true
        },
        {
          "name": "destinationToken",
          "writable": true
        },
        {
          "name": "holdPosition",
          "writable": true,
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
                "path": "sourceToken"
              }
            ]
          }
        },
        {
          "name": "holdRecord",
          "writable": true,
          "pda": {
            "seeds": [
              {
                "kind": "const",
                "value": [
                  104,
                  111,
                  108,
                  100
                ]
              },
              {
                "kind": "account",
                "path": "mint"
              },
              {
                "kind": "account",
                "path": "sourceToken"
              },
              {
                "kind": "arg",
                "path": "holdId"
              }
            ]
          }
        },
        {
          "name": "holdAuthority",
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
                  97,
                  117,
                  116,
                  104,
                  111,
                  114,
                  105,
                  116,
                  121
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
          "name": "operationsAuthority",
          "pda": {
            "seeds": [
              {
                "kind": "const",
                "value": [
                  112,
                  101,
                  114,
                  109,
                  97,
                  110,
                  101,
                  110,
                  116,
                  95,
                  100,
                  101,
                  108,
                  101,
                  103,
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
                152,
                188,
                213,
                228,
                117,
                173,
                225,
                118,
                58,
                87,
                148,
                212,
                2,
                68,
                26,
                1,
                69,
                142,
                27,
                188,
                170,
                240,
                225,
                36,
                217,
                215,
                5,
                20,
                168,
                160,
                204,
                167
              ]
            }
          }
        },
        {
          "name": "operationsProgram",
          "address": "BHDyg8PeUyVBpmkcjYLdnt3VCmYf4wp8Xeu6TXREiLKp"
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
          "name": "deployProgram",
          "address": "HCe5Um7ThFBzDSyn256EPQvyr6jy6E66ydzZ5hMta3Tq"
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
                "path": "sourceToken"
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
          "name": "destinationWhitelistPda",
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
                "path": "destinationToken"
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
          "name": "freezeProgram",
          "address": "8L1kqDvAYC9dQXNNNnZbABtRbHGjzoxSgAPzbQZmwmSd"
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
                "path": "sourceToken"
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
                "path": "sourceToken"
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
          "name": "token2022Program",
          "address": "TokenzQdBNbLqP5VEhdkAS6EPFLC1PHnBqCXEpPxuEb"
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
          "name": "holdId",
          "type": "u64"
        },
        {
          "name": "amount",
          "type": "u64"
        }
      ]
    },
    {
      "name": "reclaimHold",
      "discriminator": [
        86,
        91,
        37,
        151,
        13,
        61,
        31,
        18
      ],
      "accounts": [
        {
          "name": "caller",
          "signer": true
        },
        {
          "name": "mint",
          "docs": [
            "`hold_record` seeds, which no other mint can reproduce."
          ]
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
          "name": "tokenAccount"
        },
        {
          "name": "holdPosition",
          "writable": true,
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
                "path": "tokenAccount"
              }
            ]
          }
        },
        {
          "name": "holdRecord",
          "writable": true,
          "pda": {
            "seeds": [
              {
                "kind": "const",
                "value": [
                  104,
                  111,
                  108,
                  100
                ]
              },
              {
                "kind": "account",
                "path": "mint"
              },
              {
                "kind": "account",
                "path": "tokenAccount"
              },
              {
                "kind": "arg",
                "path": "holdId"
              }
            ]
          }
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
          "name": "holdId",
          "type": "u64"
        }
      ]
    },
    {
      "name": "releaseHold",
      "discriminator": [
        106,
        109,
        70,
        162,
        197,
        158,
        92,
        243
      ],
      "accounts": [
        {
          "name": "escrow",
          "signer": true
        },
        {
          "name": "mint",
          "docs": [
            "`hold_record` seeds, which no other mint can reproduce."
          ]
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
          "name": "tokenAccount"
        },
        {
          "name": "holdPosition",
          "writable": true,
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
                "path": "tokenAccount"
              }
            ]
          }
        },
        {
          "name": "holdRecord",
          "writable": true,
          "pda": {
            "seeds": [
              {
                "kind": "const",
                "value": [
                  104,
                  111,
                  108,
                  100
                ]
              },
              {
                "kind": "account",
                "path": "mint"
              },
              {
                "kind": "account",
                "path": "tokenAccount"
              },
              {
                "kind": "arg",
                "path": "holdId"
              }
            ]
          }
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
          "name": "holdId",
          "type": "u64"
        },
        {
          "name": "amount",
          "type": "u64"
        }
      ]
    }
  ],
  "accounts": [
    {
      "name": "hold",
      "discriminator": [
        110,
        65,
        238,
        142,
        146,
        91,
        196,
        171
      ]
    },
    {
      "name": "holdPosition",
      "discriminator": [
        126,
        32,
        74,
        45,
        205,
        98,
        8,
        28
      ]
    }
  ],
  "events": [
    {
      "name": "controllerHoldCreated",
      "discriminator": [
        245,
        54,
        4,
        209,
        43,
        149,
        224,
        225
      ]
    },
    {
      "name": "holdCreated",
      "discriminator": [
        151,
        126,
        140,
        207,
        163,
        56,
        160,
        216
      ]
    },
    {
      "name": "holdExecuted",
      "discriminator": [
        0,
        66,
        40,
        156,
        42,
        195,
        127,
        203
      ]
    },
    {
      "name": "holdReclaimed",
      "discriminator": [
        24,
        51,
        107,
        114,
        172,
        221,
        248,
        14
      ]
    },
    {
      "name": "holdReleased",
      "discriminator": [
        168,
        49,
        116,
        33,
        191,
        86,
        197,
        90
      ]
    }
  ],
  "errors": [
    {
      "code": 6000,
      "name": "zeroAmount",
      "msg": "Hold amount must be greater than zero"
    },
    {
      "code": 6001,
      "name": "expirationInThePast",
      "msg": "Hold expiration must be in the future"
    },
    {
      "code": 6002,
      "name": "holdIdMismatch",
      "msg": "Hold id does not match the next id for this position"
    },
    {
      "code": 6003,
      "name": "insufficientAvailableBalance",
      "msg": "Balance available after existing liens does not cover the hold amount"
    },
    {
      "code": 6004,
      "name": "notTheEscrow",
      "msg": "Signer is not the escrow of this hold"
    },
    {
      "code": 6005,
      "name": "holdNotActive",
      "msg": "Hold is no longer active"
    },
    {
      "code": 6006,
      "name": "holdExpired",
      "msg": "Hold has expired"
    },
    {
      "code": 6007,
      "name": "holdNotExpired",
      "msg": "Hold has not expired yet"
    },
    {
      "code": 6008,
      "name": "amountExceedsHold",
      "msg": "Amount exceeds the hold's remaining amount"
    },
    {
      "code": 6009,
      "name": "destinationMismatch",
      "msg": "Destination does not match the one pinned at hold creation"
    },
    {
      "code": 6010,
      "name": "heldAmountUnderflow",
      "msg": "Held amount is inconsistent with the hold being resolved"
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
    },
    {
      "name": "controllerHoldCreated",
      "type": {
        "kind": "struct",
        "fields": [
          {
            "name": "mint",
            "type": "pubkey"
          },
          {
            "name": "tokenAccount",
            "type": "pubkey"
          },
          {
            "name": "holdId",
            "type": "u64"
          },
          {
            "name": "controller",
            "type": "pubkey"
          },
          {
            "name": "escrow",
            "type": "pubkey"
          },
          {
            "name": "destination",
            "type": {
              "option": "pubkey"
            }
          },
          {
            "name": "amount",
            "type": "u64"
          },
          {
            "name": "expiration",
            "type": "i64"
          }
        ]
      }
    },
    {
      "name": "hold",
      "type": {
        "kind": "struct",
        "fields": [
          {
            "name": "mint",
            "type": "pubkey"
          },
          {
            "name": "tokenAccount",
            "type": "pubkey"
          },
          {
            "name": "holdId",
            "type": "u64"
          },
          {
            "name": "escrow",
            "type": "pubkey"
          },
          {
            "name": "destination",
            "type": {
              "option": "pubkey"
            }
          },
          {
            "name": "initialAmount",
            "type": "u64"
          },
          {
            "name": "currentAmount",
            "type": "u64"
          },
          {
            "name": "createdAt",
            "type": "i64"
          },
          {
            "name": "expiration",
            "type": "i64"
          },
          {
            "name": "status",
            "type": {
              "defined": {
                "name": "holdStatus"
              }
            }
          },
          {
            "name": "bump",
            "type": "u8"
          }
        ]
      }
    },
    {
      "name": "holdCreated",
      "type": {
        "kind": "struct",
        "fields": [
          {
            "name": "mint",
            "type": "pubkey"
          },
          {
            "name": "tokenAccount",
            "type": "pubkey"
          },
          {
            "name": "holdId",
            "type": "u64"
          },
          {
            "name": "escrow",
            "type": "pubkey"
          },
          {
            "name": "destination",
            "type": {
              "option": "pubkey"
            }
          },
          {
            "name": "amount",
            "type": "u64"
          },
          {
            "name": "expiration",
            "type": "i64"
          }
        ]
      }
    },
    {
      "name": "holdExecuted",
      "type": {
        "kind": "struct",
        "fields": [
          {
            "name": "mint",
            "type": "pubkey"
          },
          {
            "name": "tokenAccount",
            "type": "pubkey"
          },
          {
            "name": "holdId",
            "type": "u64"
          },
          {
            "name": "escrow",
            "type": "pubkey"
          },
          {
            "name": "destination",
            "type": "pubkey"
          },
          {
            "name": "amount",
            "type": "u64"
          },
          {
            "name": "remainingAmount",
            "type": "u64"
          }
        ]
      }
    },
    {
      "name": "holdPosition",
      "type": {
        "kind": "struct",
        "fields": [
          {
            "name": "mint",
            "type": "pubkey"
          },
          {
            "name": "tokenAccount",
            "type": "pubkey"
          },
          {
            "name": "heldAmount",
            "type": "u64"
          },
          {
            "name": "holdCount",
            "docs": [
              "Count of holds created against this position, never reset. The id",
              "assigned to the next hold is `hold_count + 1`, so the first hold gets",
              "`hold_id == 1` (mirrors `coupon`'s `count + 1` numbering)."
            ],
            "type": "u64"
          },
          {
            "name": "bump",
            "type": "u8"
          }
        ]
      }
    },
    {
      "name": "holdReclaimed",
      "type": {
        "kind": "struct",
        "fields": [
          {
            "name": "mint",
            "type": "pubkey"
          },
          {
            "name": "tokenAccount",
            "type": "pubkey"
          },
          {
            "name": "holdId",
            "type": "u64"
          },
          {
            "name": "caller",
            "type": "pubkey"
          },
          {
            "name": "amount",
            "type": "u64"
          }
        ]
      }
    },
    {
      "name": "holdReleased",
      "type": {
        "kind": "struct",
        "fields": [
          {
            "name": "mint",
            "type": "pubkey"
          },
          {
            "name": "tokenAccount",
            "type": "pubkey"
          },
          {
            "name": "holdId",
            "type": "u64"
          },
          {
            "name": "escrow",
            "type": "pubkey"
          },
          {
            "name": "amount",
            "type": "u64"
          },
          {
            "name": "remainingAmount",
            "type": "u64"
          }
        ]
      }
    },
    {
      "name": "holdStatus",
      "type": {
        "kind": "enum",
        "variants": [
          {
            "name": "active"
          },
          {
            "name": "expired"
          },
          {
            "name": "closed"
          }
        ]
      }
    }
  ]
};
