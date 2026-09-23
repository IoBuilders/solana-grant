/**
 * Program IDL in camelCase format in order to be used in JS/TS.
 *
 * Note that this is only a type helper and is not the actual IDL. The original
 * IDL can be found at `target/idl/coupon.json`.
 */
export type Coupon = {
  "address": "CGQMgamBMtJ97CCMwVD9v5vAYVzFsXLy8beN8Ej6t3FK",
  "metadata": {
    "name": "coupon",
    "version": "0.1.0",
    "spec": "0.1.0",
    "description": "Created with Anchor"
  },
  "instructions": [
    {
      "name": "createCoupon",
      "discriminator": [
        29,
        170,
        159,
        88,
        211,
        20,
        13,
        56
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
          "name": "couponAuthority",
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
          "name": "couponCounter",
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
          "name": "coupon",
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
            ]
          }
        },
        {
          "name": "snapshotCounter",
          "docs": [
            "validated inside `take_snapshot`."
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
          "name": "snapshotMerkleRoot",
          "writable": true
        },
        {
          "name": "snapshotProgram",
          "address": "hgUtrpstViwxutrkoVXwQh3GQC18wHAmuAvYFTNiV2M"
        },
        {
          "name": "systemProgram",
          "address": "11111111111111111111111111111111"
        },
        {
          "name": "snapshotEventAuthority"
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
          "name": "couponId",
          "type": "u64"
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
    },
    {
      "name": "setCouponRate",
      "discriminator": [
        163,
        75,
        145,
        124,
        107,
        232,
        224,
        13
      ],
      "accounts": [
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
          "name": "coupon",
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
          "name": "interestRate",
          "type": {
            "option": "u64"
          }
        },
        {
          "name": "interestRateDecimals",
          "type": {
            "option": "u8"
          }
        }
      ]
    }
  ],
  "accounts": [
    {
      "name": "coupon",
      "discriminator": [
        24,
        230,
        224,
        210,
        200,
        206,
        79,
        57
      ]
    },
    {
      "name": "couponCounter",
      "discriminator": [
        204,
        228,
        141,
        100,
        221,
        9,
        198,
        67
      ]
    }
  ],
  "events": [
    {
      "name": "couponCreated",
      "discriminator": [
        11,
        158,
        13,
        126,
        64,
        79,
        194,
        48
      ]
    },
    {
      "name": "couponRateSet",
      "discriminator": [
        95,
        144,
        111,
        94,
        94,
        94,
        15,
        33
      ]
    }
  ],
  "errors": [
    {
      "code": 6000,
      "name": "invalidCouponId",
      "msg": "Supplied coupon_id does not match coupon_counter.count + 1"
    },
    {
      "code": 6001,
      "name": "invalidCouponPeriod",
      "msg": "period_end_date must be strictly greater than period_start_date"
    },
    {
      "code": 6002,
      "name": "invalidPaymentDate",
      "msg": "payment_date must be strictly greater than period_end_date"
    },
    {
      "code": 6003,
      "name": "inconsistentRateOverride",
      "msg": "interest_rate_override and interest_rate_override_decimals must both be Some or both be None"
    },
    {
      "code": 6004,
      "name": "couponCounterOverflow",
      "msg": "coupon counter overflow when creating new coupon"
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
      "name": "couponCounter",
      "type": {
        "kind": "struct",
        "fields": [
          {
            "name": "bump",
            "type": "u8"
          },
          {
            "name": "count",
            "type": "u64"
          }
        ]
      }
    },
    {
      "name": "couponCreated",
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
      "name": "couponRateSet",
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
    }
  ]
};
