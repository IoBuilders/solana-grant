/**
 * Program IDL in camelCase format in order to be used in JS/TS.
 *
 * Note that this is only a type helper and is not the actual IDL. The original
 * IDL can be found at `target/idl/bond.json`.
 */
export type Bond = {
  "address": "8opYXiWzWBrUEr5vtcvaX1ybzYaMKrndxkW1U9Patk46",
  "metadata": {
    "name": "bond",
    "version": "0.1.0",
    "spec": "0.1.0",
    "description": "Created with Anchor"
  },
  "instructions": [
    {
      "name": "updateBondTerms",
      "discriminator": [
        154,
        161,
        195,
        201,
        104,
        189,
        48,
        239
      ],
      "accounts": [
        {
          "name": "payer",
          "writable": true,
          "signer": true
        },
        {
          "name": "authority",
          "writable": true,
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
          "name": "bondTerms",
          "writable": true,
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
          "name": "args",
          "type": {
            "defined": {
              "name": "bondTermsArgs"
            }
          }
        }
      ]
    }
  ],
  "accounts": [
    {
      "name": "bondTerms",
      "discriminator": [
        47,
        124,
        108,
        80,
        69,
        221,
        184,
        149
      ]
    }
  ],
  "events": [
    {
      "name": "bondTermsUpdated",
      "discriminator": [
        86,
        138,
        186,
        99,
        75,
        46,
        94,
        106
      ]
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
      "name": "bondTermsArgs",
      "type": {
        "kind": "struct",
        "fields": [
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
      "name": "bondTermsUpdated",
      "type": {
        "kind": "struct",
        "fields": [
          {
            "name": "mint",
            "type": "pubkey"
          },
          {
            "name": "operator",
            "type": "pubkey"
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
    }
  ]
};
