/**
 * Program IDL in camelCase format in order to be used in JS/TS.
 *
 * Note that this is only a type helper and is not the actual IDL. The original
 * IDL can be found at `target/idl/deploy.json`.
 */
export type Deploy = {
  "address": "HCe5Um7ThFBzDSyn256EPQvyr6jy6E66ydzZ5hMta3Tq",
  "metadata": {
    "name": "deploy",
    "version": "0.1.0",
    "spec": "0.1.0",
    "description": "Created with Anchor"
  },
  "instructions": [
    {
      "name": "deployMint",
      "discriminator": [
        22,
        164,
        150,
        116,
        55,
        220,
        114,
        111
      ],
      "accounts": [
        {
          "name": "payer",
          "writable": true,
          "signer": true
        },
        {
          "name": "deployer",
          "signer": true
        },
        {
          "name": "assetConfigurationPda",
          "writable": true,
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
            ]
          }
        },
        {
          "name": "mint",
          "docs": [
            "the extension and mint initialization CPIs below."
          ],
          "writable": true,
          "signer": true
        },
        {
          "name": "tempMintAuthority",
          "pda": {
            "seeds": [
              {
                "kind": "const",
                "value": [
                  116,
                  101,
                  109,
                  112,
                  95,
                  109,
                  105,
                  110,
                  116,
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
          "name": "mintAuthority",
          "pda": {
            "seeds": [
              {
                "kind": "const",
                "value": [
                  109,
                  105,
                  110,
                  116,
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
            ],
            "program": {
              "kind": "const",
              "value": [
                158,
                179,
                55,
                82,
                46,
                86,
                144,
                48,
                111,
                202,
                27,
                236,
                31,
                22,
                130,
                134,
                127,
                156,
                178,
                206,
                206,
                46,
                92,
                143,
                45,
                125,
                106,
                75,
                191,
                170,
                176,
                163
              ]
            }
          }
        },
        {
          "name": "permanentDelegateAuthority",
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
          "name": "permissionedBurnAuthority",
          "pda": {
            "seeds": [
              {
                "kind": "const",
                "value": [
                  112,
                  101,
                  114,
                  109,
                  105,
                  115,
                  115,
                  105,
                  111,
                  110,
                  101,
                  100,
                  95,
                  98,
                  117,
                  114,
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
          "name": "metadataUpdateAuthority",
          "pda": {
            "seeds": [
              {
                "kind": "const",
                "value": [
                  109,
                  101,
                  116,
                  97,
                  100,
                  97,
                  116,
                  97,
                  95,
                  117,
                  112,
                  100,
                  97,
                  116,
                  101,
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
            ],
            "program": {
              "kind": "const",
              "value": [
                10,
                157,
                223,
                137,
                180,
                23,
                231,
                223,
                245,
                31,
                117,
                191,
                188,
                253,
                245,
                6,
                4,
                75,
                11,
                77,
                30,
                157,
                149,
                105,
                74,
                5,
                250,
                203,
                71,
                61,
                106,
                84
              ]
            }
          }
        },
        {
          "name": "pausableAuthority",
          "pda": {
            "seeds": [
              {
                "kind": "const",
                "value": [
                  112,
                  97,
                  117,
                  115,
                  97,
                  98,
                  108,
                  101,
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
            ],
            "program": {
              "kind": "const",
              "value": [
                70,
                51,
                173,
                173,
                71,
                175,
                47,
                136,
                251,
                177,
                216,
                40,
                156,
                189,
                36,
                196,
                237,
                82,
                161,
                156,
                44,
                172,
                227,
                117,
                105,
                243,
                72,
                234,
                72,
                173,
                11,
                52
              ]
            }
          }
        },
        {
          "name": "transferHookAuthority",
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
                  104,
                  111,
                  111,
                  107,
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
          "name": "extraAccountMetaList",
          "writable": true,
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
          "name": "token2022Program",
          "address": "TokenzQdBNbLqP5VEhdkAS6EPFLC1PHnBqCXEpPxuEb"
        },
        {
          "name": "systemProgram",
          "address": "11111111111111111111111111111111"
        },
        {
          "name": "rent",
          "address": "SysvarRent111111111111111111111111111111111"
        },
        {
          "name": "rolesPda",
          "writable": true,
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
                "path": "deployer"
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
          "name": "accessControlProgram",
          "address": "GpyjQqBWux3JYqxKCXFrDbWZmhFWBJWVaVivkBW2DL2w"
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
          "name": "params",
          "type": {
            "defined": {
              "name": "deployMintParams"
            }
          }
        }
      ]
    }
  ],
  "accounts": [
    {
      "name": "assetConfiguration",
      "discriminator": [
        15,
        79,
        132,
        40,
        8,
        129,
        114,
        149
      ]
    }
  ],
  "events": [
    {
      "name": "mintDeployed",
      "discriminator": [
        131,
        178,
        106,
        183,
        102,
        248,
        105,
        167
      ]
    }
  ],
  "errors": [
    {
      "code": 6000,
      "name": "mintAuthorityMustBeSigner",
      "msg": "Mint authority must be a signer"
    },
    {
      "code": 6001,
      "name": "invalidMintAccountSize",
      "msg": "Failed to calculate mint account size for the requested extensions"
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
      "name": "deployMintParams",
      "type": {
        "kind": "struct",
        "fields": [
          {
            "name": "decimals",
            "type": "u8"
          },
          {
            "name": "name",
            "type": "string"
          },
          {
            "name": "symbol",
            "type": "string"
          },
          {
            "name": "uri",
            "type": "string"
          },
          {
            "name": "additionalMetadata",
            "type": {
              "vec": {
                "defined": {
                  "name": "metadataField"
                }
              }
            }
          },
          {
            "name": "assetClassConfigId",
            "type": "u64"
          },
          {
            "name": "assetClassVersionId",
            "type": "u64"
          }
        ]
      }
    },
    {
      "name": "metadataField",
      "type": {
        "kind": "struct",
        "fields": [
          {
            "name": "key",
            "type": "string"
          },
          {
            "name": "value",
            "type": "string"
          }
        ]
      }
    },
    {
      "name": "mintDeployed",
      "type": {
        "kind": "struct",
        "fields": [
          {
            "name": "mint",
            "type": "pubkey"
          },
          {
            "name": "deployer",
            "type": "pubkey"
          },
          {
            "name": "decimals",
            "type": "u8"
          },
          {
            "name": "name",
            "type": "string"
          },
          {
            "name": "symbol",
            "type": "string"
          },
          {
            "name": "uri",
            "type": "string"
          },
          {
            "name": "isin",
            "type": {
              "option": "string"
            }
          },
          {
            "name": "assetClassConfigId",
            "type": "u64"
          },
          {
            "name": "assetClassVersionId",
            "type": "u64"
          }
        ]
      }
    }
  ]
};
