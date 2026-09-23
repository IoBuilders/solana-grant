/**
 * Program IDL in camelCase format in order to be used in JS/TS.
 *
 * Note that this is only a type helper and is not the actual IDL. The original
 * IDL can be found at `target/idl/treasury.json`.
 */
export type Treasury = {
  "address": "G71RRNtr2PLZ9Tbmp9CKnxghf3aMoasUwLGPb2u7BytA",
  "metadata": {
    "name": "treasury",
    "version": "0.1.0",
    "spec": "0.1.0",
    "description": "Created with Anchor"
  },
  "instructions": [
    {
      "name": "payCoupon",
      "discriminator": [
        235,
        226,
        45,
        115,
        95,
        34,
        46,
        110
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
          "name": "treasuryConfig",
          "writable": true,
          "pda": {
            "seeds": [
              {
                "kind": "const",
                "value": [
                  116,
                  114,
                  101,
                  97,
                  115,
                  117,
                  114,
                  121,
                  95,
                  99,
                  111,
                  110,
                  102,
                  105,
                  103
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
          "name": "treasuryAuthority",
          "pda": {
            "seeds": [
              {
                "kind": "const",
                "value": [
                  116,
                  114,
                  101,
                  97,
                  115,
                  117,
                  114,
                  121,
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
          "name": "paymentMint"
        },
        {
          "name": "treasuryTokenAccount",
          "writable": true
        },
        {
          "name": "holderPaymentAccount",
          "writable": true
        },
        {
          "name": "bondTerms",
          "pda": {
            "seeds": [
              {
                "kind": "const",
                "value": [
                  98,
                  111,
                  110,
                  100,
                  95,
                  116,
                  101,
                  114,
                  109,
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
                116,
                0,
                72,
                143,
                85,
                49,
                115,
                234,
                66,
                223,
                13,
                53,
                47,
                220,
                177,
                250,
                156,
                201,
                51,
                16,
                149,
                156,
                39,
                140,
                253,
                210,
                82,
                254,
                7,
                250,
                186,
                7
              ]
            }
          }
        },
        {
          "name": "coupon",
          "pda": {
            "seeds": [
              {
                "kind": "const",
                "value": [
                  99,
                  111,
                  117,
                  112,
                  111,
                  110
                ]
              },
              {
                "kind": "account",
                "path": "mint"
              },
              {
                "kind": "arg",
                "path": "couponId"
              }
            ],
            "program": {
              "kind": "const",
              "value": [
                167,
                98,
                173,
                126,
                119,
                180,
                215,
                106,
                2,
                46,
                241,
                208,
                13,
                86,
                8,
                91,
                213,
                185,
                63,
                97,
                157,
                152,
                222,
                167,
                130,
                194,
                158,
                201,
                73,
                76,
                82,
                206
              ]
            }
          }
        },
        {
          "name": "snapshotMerkleRoot",
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
                "path": "coupon.snapshot_id",
                "account": "coupon"
              }
            ],
            "program": {
              "kind": "const",
              "value": [
                10,
                107,
                223,
                164,
                3,
                187,
                104,
                46,
                78,
                61,
                30,
                172,
                109,
                69,
                28,
                224,
                240,
                205,
                168,
                113,
                10,
                68,
                9,
                225,
                67,
                66,
                92,
                96,
                105,
                114,
                65,
                54
              ]
            }
          }
        },
        {
          "name": "couponPaid",
          "writable": true,
          "pda": {
            "seeds": [
              {
                "kind": "const",
                "value": [
                  99,
                  111,
                  117,
                  112,
                  111,
                  110,
                  95,
                  112,
                  97,
                  105,
                  100
                ]
              },
              {
                "kind": "account",
                "path": "mint"
              },
              {
                "kind": "arg",
                "path": "couponId"
              },
              {
                "kind": "arg",
                "path": "account"
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
          "name": "eventAuthority"
        },
        {
          "name": "program"
        }
      ],
      "args": [
        {
          "name": "couponId",
          "type": "u64"
        },
        {
          "name": "account",
          "type": "pubkey"
        },
        {
          "name": "balance",
          "type": "u64"
        },
        {
          "name": "merkleProof",
          "type": {
            "vec": {
              "array": [
                "u8",
                32
              ]
            }
          }
        }
      ]
    },
    {
      "name": "setPaymentToken",
      "discriminator": [
        155,
        213,
        140,
        249,
        53,
        59,
        20,
        5
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
          "name": "treasuryConfig",
          "writable": true,
          "pda": {
            "seeds": [
              {
                "kind": "const",
                "value": [
                  116,
                  114,
                  101,
                  97,
                  115,
                  117,
                  114,
                  121,
                  95,
                  99,
                  111,
                  110,
                  102,
                  105,
                  103
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
          "name": "couponCounter",
          "pda": {
            "seeds": [
              {
                "kind": "const",
                "value": [
                  99,
                  111,
                  117,
                  112,
                  111,
                  110,
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
            ],
            "program": {
              "kind": "const",
              "value": [
                167,
                98,
                173,
                126,
                119,
                180,
                215,
                106,
                2,
                46,
                241,
                208,
                13,
                86,
                8,
                91,
                213,
                185,
                63,
                97,
                157,
                152,
                222,
                167,
                130,
                194,
                158,
                201,
                73,
                76,
                82,
                206
              ]
            }
          }
        },
        {
          "name": "paymentMint"
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
          "name": "systemProgram",
          "address": "11111111111111111111111111111111"
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
          "name": "eventAuthority"
        },
        {
          "name": "program"
        }
      ],
      "args": []
    }
  ],
  "accounts": [
    {
      "name": "couponPaidMarker",
      "discriminator": [
        190,
        24,
        251,
        23,
        150,
        191,
        73,
        186
      ]
    },
    {
      "name": "treasuryConfig",
      "discriminator": [
        124,
        54,
        212,
        227,
        213,
        189,
        168,
        41
      ]
    }
  ],
  "events": [
    {
      "name": "couponPaid",
      "discriminator": [
        11,
        228,
        86,
        67,
        111,
        22,
        107,
        226
      ]
    },
    {
      "name": "paymentTokenSet",
      "discriminator": [
        235,
        37,
        70,
        239,
        23,
        99,
        68,
        232
      ]
    }
  ],
  "errors": [
    {
      "code": 6000,
      "name": "negativeElapsedTime",
      "msg": "Coupon payment_date is earlier than bond issuance_date"
    },
    {
      "code": 6001,
      "name": "amountOverflow",
      "msg": "Computed coupon amount overflows u64"
    },
    {
      "code": 6002,
      "name": "couponNotMature",
      "msg": "Coupon payment_date has not been reached yet"
    },
    {
      "code": 6003,
      "name": "claimsInProgress",
      "msg": "Payment token cannot be changed while claims are in progress for the current coupon"
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
      "name": "bondTerms",
      "type": {
        "kind": "struct",
        "fields": [
          {
            "name": "bump",
            "type": "u8"
          },
          {
            "name": "interestRate",
            "type": "u64"
          },
          {
            "name": "interestRateDecimals",
            "type": "u8"
          },
          {
            "name": "parValue",
            "type": "u64"
          },
          {
            "name": "parValueDecimals",
            "type": "u8"
          },
          {
            "name": "minimumDenomination",
            "type": "u64"
          },
          {
            "name": "issuanceDate",
            "type": "i64"
          },
          {
            "name": "dayCountConvention",
            "type": {
              "defined": {
                "name": "dayCountConvention"
              }
            }
          }
        ]
      }
    },
    {
      "name": "coupon",
      "type": {
        "kind": "struct",
        "fields": [
          {
            "name": "bump",
            "type": "u8"
          },
          {
            "name": "snapshotId",
            "type": "u64"
          },
          {
            "name": "periodStartDate",
            "type": "i64"
          },
          {
            "name": "periodEndDate",
            "type": "i64"
          },
          {
            "name": "paymentDate",
            "type": "i64"
          },
          {
            "name": "interestRateOverride",
            "type": {
              "option": "u64"
            }
          },
          {
            "name": "interestRateOverrideDecimals",
            "type": {
              "option": "u8"
            }
          }
        ]
      }
    },
    {
      "name": "couponPaid",
      "type": {
        "kind": "struct",
        "fields": [
          {
            "name": "mint",
            "type": "pubkey"
          },
          {
            "name": "couponId",
            "type": "u64"
          },
          {
            "name": "holderTokenAccount",
            "type": "pubkey"
          },
          {
            "name": "paymentMint",
            "type": "pubkey"
          },
          {
            "name": "amount",
            "type": "u64"
          },
          {
            "name": "payer",
            "type": "pubkey"
          }
        ]
      }
    },
    {
      "name": "couponPaidMarker",
      "type": {
        "kind": "struct",
        "fields": [
          {
            "name": "bump",
            "type": "u8"
          },
          {
            "name": "amount",
            "type": "u64"
          }
        ]
      }
    },
    {
      "name": "dayCountConvention",
      "type": {
        "kind": "enum",
        "variants": [
          {
            "name": "actual360"
          },
          {
            "name": "actual365"
          }
        ]
      }
    },
    {
      "name": "paymentTokenSet",
      "type": {
        "kind": "struct",
        "fields": [
          {
            "name": "mint",
            "type": "pubkey"
          },
          {
            "name": "paymentMint",
            "type": "pubkey"
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
      "name": "treasuryConfig",
      "type": {
        "kind": "struct",
        "fields": [
          {
            "name": "bump",
            "type": "u8"
          },
          {
            "name": "paymentMint",
            "type": "pubkey"
          },
          {
            "name": "paymentMintDecimals",
            "type": "u8"
          },
          {
            "name": "lockedForCouponId",
            "type": "u64"
          }
        ]
      }
    }
  ]
};
