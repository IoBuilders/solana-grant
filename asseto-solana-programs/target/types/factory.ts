/**
 * Program IDL in camelCase format in order to be used in JS/TS.
 *
 * Note that this is only a type helper and is not the actual IDL. The original
 * IDL can be found at `target/idl/factory.json`.
 */
export type Factory = {
  "address": "FEY9E77nH7R1gLGNxkhYKchJpB6MgpMrWMhkNXrNhzR5",
  "metadata": {
    "name": "factory",
    "version": "0.1.0",
    "spec": "0.1.0",
    "description": "Created with Anchor"
  },
  "instructions": [
    {
      "name": "acceptAssetClassOwnership",
      "discriminator": [
        205,
        253,
        86,
        237,
        39,
        143,
        0,
        10
      ],
      "accounts": [
        {
          "name": "pendingOwner",
          "writable": true,
          "signer": true
        },
        {
          "name": "factory",
          "pda": {
            "seeds": [
              {
                "kind": "const",
                "value": [
                  102,
                  97,
                  99,
                  116,
                  111,
                  114,
                  121
                ]
              }
            ]
          }
        },
        {
          "name": "assetClassOwnershipPda",
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
                  108,
                  97,
                  115,
                  115,
                  95,
                  111,
                  119,
                  110,
                  101,
                  114,
                  115,
                  104,
                  105,
                  112
                ]
              },
              {
                "kind": "arg",
                "path": "configId"
              }
            ]
          }
        },
        {
          "name": "assetClassPendingOwnerPda",
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
                  108,
                  97,
                  115,
                  115,
                  95,
                  112,
                  101,
                  110,
                  100,
                  105,
                  110,
                  103,
                  95,
                  111,
                  119,
                  110,
                  101,
                  114
                ]
              },
              {
                "kind": "arg",
                "path": "configId"
              }
            ]
          }
        }
      ],
      "args": [
        {
          "name": "configId",
          "type": "u64"
        }
      ]
    },
    {
      "name": "acceptNomination",
      "discriminator": [
        126,
        254,
        37,
        185,
        168,
        32,
        50,
        143
      ],
      "accounts": [
        {
          "name": "pendingManager",
          "writable": true,
          "signer": true
        },
        {
          "name": "factory",
          "writable": true,
          "pda": {
            "seeds": [
              {
                "kind": "const",
                "value": [
                  102,
                  97,
                  99,
                  116,
                  111,
                  114,
                  121
                ]
              }
            ]
          }
        },
        {
          "name": "factoryPendingManagerPda",
          "writable": true,
          "pda": {
            "seeds": [
              {
                "kind": "const",
                "value": [
                  102,
                  97,
                  99,
                  116,
                  111,
                  114,
                  121,
                  95,
                  112,
                  101,
                  110,
                  100,
                  105,
                  110,
                  103,
                  95,
                  109,
                  97,
                  110,
                  97,
                  103,
                  101,
                  114
                ]
              }
            ]
          }
        }
      ],
      "args": []
    },
    {
      "name": "cancelAssetClassOwnership",
      "discriminator": [
        71,
        143,
        99,
        79,
        221,
        192,
        153,
        186
      ],
      "accounts": [
        {
          "name": "currentOwner",
          "writable": true,
          "signer": true
        },
        {
          "name": "factory",
          "pda": {
            "seeds": [
              {
                "kind": "const",
                "value": [
                  102,
                  97,
                  99,
                  116,
                  111,
                  114,
                  121
                ]
              }
            ]
          }
        },
        {
          "name": "assetClassOwnershipPda",
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
                  111,
                  119,
                  110,
                  101,
                  114,
                  115,
                  104,
                  105,
                  112
                ]
              },
              {
                "kind": "arg",
                "path": "configId"
              }
            ]
          }
        },
        {
          "name": "assetClassPendingOwnerPda",
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
                  108,
                  97,
                  115,
                  115,
                  95,
                  112,
                  101,
                  110,
                  100,
                  105,
                  110,
                  103,
                  95,
                  111,
                  119,
                  110,
                  101,
                  114
                ]
              },
              {
                "kind": "arg",
                "path": "configId"
              }
            ]
          }
        }
      ],
      "args": [
        {
          "name": "configId",
          "type": "u64"
        }
      ]
    },
    {
      "name": "cancelNomination",
      "discriminator": [
        95,
        2,
        136,
        89,
        181,
        232,
        153,
        174
      ],
      "accounts": [
        {
          "name": "currentManager",
          "writable": true,
          "signer": true
        },
        {
          "name": "factory",
          "pda": {
            "seeds": [
              {
                "kind": "const",
                "value": [
                  102,
                  97,
                  99,
                  116,
                  111,
                  114,
                  121
                ]
              }
            ]
          }
        },
        {
          "name": "factoryPendingManagerPda",
          "writable": true,
          "pda": {
            "seeds": [
              {
                "kind": "const",
                "value": [
                  102,
                  97,
                  99,
                  116,
                  111,
                  114,
                  121,
                  95,
                  112,
                  101,
                  110,
                  100,
                  105,
                  110,
                  103,
                  95,
                  109,
                  97,
                  110,
                  97,
                  103,
                  101,
                  114
                ]
              }
            ]
          }
        }
      ],
      "args": []
    },
    {
      "name": "createAssetClass",
      "discriminator": [
        54,
        43,
        221,
        168,
        187,
        80,
        204,
        18
      ],
      "accounts": [
        {
          "name": "manager",
          "writable": true,
          "signer": true
        },
        {
          "name": "factory",
          "pda": {
            "seeds": [
              {
                "kind": "const",
                "value": [
                  102,
                  97,
                  99,
                  116,
                  111,
                  114,
                  121
                ]
              }
            ]
          }
        },
        {
          "name": "assetClassOwnershipPda",
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
                  108,
                  97,
                  115,
                  115,
                  95,
                  111,
                  119,
                  110,
                  101,
                  114,
                  115,
                  104,
                  105,
                  112
                ]
              },
              {
                "kind": "arg",
                "path": "configId"
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
          "name": "configId",
          "type": "u64"
        },
        {
          "name": "owner",
          "type": "pubkey"
        }
      ]
    },
    {
      "name": "disableAssetClassVersionFunctionalities",
      "discriminator": [
        132,
        142,
        11,
        174,
        163,
        48,
        69,
        188
      ],
      "accounts": [
        {
          "name": "owner",
          "signer": true
        },
        {
          "name": "factory",
          "pda": {
            "seeds": [
              {
                "kind": "const",
                "value": [
                  102,
                  97,
                  99,
                  116,
                  111,
                  114,
                  121
                ]
              }
            ]
          }
        },
        {
          "name": "assetClassOwnershipPda",
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
                  111,
                  119,
                  110,
                  101,
                  114,
                  115,
                  104,
                  105,
                  112
                ]
              },
              {
                "kind": "arg",
                "path": "configId"
              }
            ]
          }
        },
        {
          "name": "assetClassVersionPda",
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
                "kind": "arg",
                "path": "configId"
              },
              {
                "kind": "arg",
                "path": "version"
              }
            ]
          }
        }
      ],
      "args": [
        {
          "name": "configId",
          "type": "u64"
        },
        {
          "name": "version",
          "type": "u64"
        },
        {
          "name": "functionalities",
          "type": {
            "vec": "u16"
          }
        }
      ]
    },
    {
      "name": "enableAssetClassVersionFunctionalities",
      "discriminator": [
        164,
        68,
        237,
        85,
        52,
        244,
        71,
        240
      ],
      "accounts": [
        {
          "name": "owner",
          "signer": true
        },
        {
          "name": "factory",
          "pda": {
            "seeds": [
              {
                "kind": "const",
                "value": [
                  102,
                  97,
                  99,
                  116,
                  111,
                  114,
                  121
                ]
              }
            ]
          }
        },
        {
          "name": "assetClassOwnershipPda",
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
                  111,
                  119,
                  110,
                  101,
                  114,
                  115,
                  104,
                  105,
                  112
                ]
              },
              {
                "kind": "arg",
                "path": "configId"
              }
            ]
          }
        },
        {
          "name": "assetClassVersionPda",
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
                "kind": "arg",
                "path": "configId"
              },
              {
                "kind": "arg",
                "path": "version"
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
          "name": "configId",
          "type": "u64"
        },
        {
          "name": "version",
          "type": "u64"
        },
        {
          "name": "functionalities",
          "type": {
            "vec": "u16"
          }
        }
      ]
    },
    {
      "name": "finalizeAssetClassVersion",
      "discriminator": [
        175,
        175,
        18,
        111,
        214,
        108,
        212,
        218
      ],
      "accounts": [
        {
          "name": "owner",
          "signer": true
        },
        {
          "name": "factory",
          "pda": {
            "seeds": [
              {
                "kind": "const",
                "value": [
                  102,
                  97,
                  99,
                  116,
                  111,
                  114,
                  121
                ]
              }
            ]
          }
        },
        {
          "name": "assetClassOwnershipPda",
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
                  108,
                  97,
                  115,
                  115,
                  95,
                  111,
                  119,
                  110,
                  101,
                  114,
                  115,
                  104,
                  105,
                  112
                ]
              },
              {
                "kind": "arg",
                "path": "configId"
              }
            ]
          }
        },
        {
          "name": "assetClassVersionPda",
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
                "kind": "arg",
                "path": "configId"
              },
              {
                "kind": "arg",
                "path": "version"
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
          "name": "configId",
          "type": "u64"
        },
        {
          "name": "version",
          "type": "u64"
        }
      ]
    },
    {
      "name": "initAssetClassVersion",
      "discriminator": [
        197,
        232,
        31,
        11,
        25,
        116,
        94,
        121
      ],
      "accounts": [
        {
          "name": "owner",
          "writable": true,
          "signer": true
        },
        {
          "name": "factory",
          "pda": {
            "seeds": [
              {
                "kind": "const",
                "value": [
                  102,
                  97,
                  99,
                  116,
                  111,
                  114,
                  121
                ]
              }
            ]
          }
        },
        {
          "name": "assetClassOwnershipPda",
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
                  111,
                  119,
                  110,
                  101,
                  114,
                  115,
                  104,
                  105,
                  112
                ]
              },
              {
                "kind": "arg",
                "path": "configId"
              }
            ]
          }
        },
        {
          "name": "assetClassVersionPda",
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
                "kind": "arg",
                "path": "configId"
              },
              {
                "kind": "arg",
                "path": "version"
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
          "name": "configId",
          "type": "u64"
        },
        {
          "name": "version",
          "type": "u64"
        }
      ]
    },
    {
      "name": "initialize",
      "discriminator": [
        175,
        175,
        109,
        31,
        13,
        152,
        155,
        237
      ],
      "accounts": [
        {
          "name": "payer",
          "writable": true,
          "signer": true
        },
        {
          "name": "manager",
          "signer": true
        },
        {
          "name": "factory",
          "writable": true,
          "pda": {
            "seeds": [
              {
                "kind": "const",
                "value": [
                  102,
                  97,
                  99,
                  116,
                  111,
                  114,
                  121
                ]
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
      "args": []
    },
    {
      "name": "nominateAssetClassOwner",
      "discriminator": [
        114,
        111,
        83,
        48,
        225,
        82,
        200,
        17
      ],
      "accounts": [
        {
          "name": "currentOwner",
          "writable": true,
          "signer": true
        },
        {
          "name": "factory",
          "pda": {
            "seeds": [
              {
                "kind": "const",
                "value": [
                  102,
                  97,
                  99,
                  116,
                  111,
                  114,
                  121
                ]
              }
            ]
          }
        },
        {
          "name": "assetClassOwnershipPda",
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
                  111,
                  119,
                  110,
                  101,
                  114,
                  115,
                  104,
                  105,
                  112
                ]
              },
              {
                "kind": "arg",
                "path": "configId"
              }
            ]
          }
        },
        {
          "name": "assetClassPendingOwnerPda",
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
                  108,
                  97,
                  115,
                  115,
                  95,
                  112,
                  101,
                  110,
                  100,
                  105,
                  110,
                  103,
                  95,
                  111,
                  119,
                  110,
                  101,
                  114
                ]
              },
              {
                "kind": "arg",
                "path": "configId"
              }
            ]
          }
        },
        {
          "name": "systemProgram",
          "address": "11111111111111111111111111111111"
        }
      ],
      "args": [
        {
          "name": "configId",
          "type": "u64"
        },
        {
          "name": "newOwner",
          "type": "pubkey"
        }
      ]
    },
    {
      "name": "nominateManager",
      "discriminator": [
        42,
        37,
        5,
        142,
        200,
        167,
        25,
        14
      ],
      "accounts": [
        {
          "name": "currentManager",
          "writable": true,
          "signer": true
        },
        {
          "name": "factory",
          "pda": {
            "seeds": [
              {
                "kind": "const",
                "value": [
                  102,
                  97,
                  99,
                  116,
                  111,
                  114,
                  121
                ]
              }
            ]
          }
        },
        {
          "name": "factoryPendingManagerPda",
          "writable": true,
          "pda": {
            "seeds": [
              {
                "kind": "const",
                "value": [
                  102,
                  97,
                  99,
                  116,
                  111,
                  114,
                  121,
                  95,
                  112,
                  101,
                  110,
                  100,
                  105,
                  110,
                  103,
                  95,
                  109,
                  97,
                  110,
                  97,
                  103,
                  101,
                  114
                ]
              }
            ]
          }
        },
        {
          "name": "systemProgram",
          "address": "11111111111111111111111111111111"
        }
      ],
      "args": [
        {
          "name": "newManager",
          "type": "pubkey"
        }
      ]
    },
    {
      "name": "pause",
      "discriminator": [
        211,
        22,
        221,
        251,
        74,
        121,
        193,
        47
      ],
      "accounts": [
        {
          "name": "manager",
          "signer": true
        },
        {
          "name": "factory",
          "writable": true,
          "pda": {
            "seeds": [
              {
                "kind": "const",
                "value": [
                  102,
                  97,
                  99,
                  116,
                  111,
                  114,
                  121
                ]
              }
            ]
          }
        }
      ],
      "args": []
    },
    {
      "name": "unpause",
      "discriminator": [
        169,
        144,
        4,
        38,
        10,
        141,
        188,
        255
      ],
      "accounts": [
        {
          "name": "manager",
          "signer": true
        },
        {
          "name": "factory",
          "writable": true,
          "pda": {
            "seeds": [
              {
                "kind": "const",
                "value": [
                  102,
                  97,
                  99,
                  116,
                  111,
                  114,
                  121
                ]
              }
            ]
          }
        }
      ],
      "args": []
    }
  ],
  "accounts": [
    {
      "name": "assetClassOwnership",
      "discriminator": [
        172,
        104,
        208,
        239,
        210,
        77,
        128,
        255
      ]
    },
    {
      "name": "assetClassPendingOwner",
      "discriminator": [
        119,
        144,
        247,
        116,
        188,
        185,
        96,
        17
      ]
    },
    {
      "name": "assetClassVersion",
      "discriminator": [
        255,
        193,
        180,
        87,
        186,
        245,
        78,
        199
      ]
    },
    {
      "name": "factory",
      "discriminator": [
        159,
        68,
        192,
        61,
        48,
        249,
        216,
        202
      ]
    },
    {
      "name": "factoryPendingManager",
      "discriminator": [
        252,
        131,
        28,
        172,
        212,
        216,
        174,
        108
      ]
    }
  ],
  "events": [
    {
      "name": "assetClassCreated",
      "discriminator": [
        62,
        173,
        188,
        91,
        122,
        94,
        47,
        172
      ]
    },
    {
      "name": "assetClassVersionFinalized",
      "discriminator": [
        116,
        56,
        203,
        0,
        55,
        146,
        124,
        87
      ]
    },
    {
      "name": "assetClassVersionFunctionalitiesEnabled",
      "discriminator": [
        40,
        130,
        149,
        165,
        60,
        25,
        94,
        79
      ]
    },
    {
      "name": "assetClassVersionInitialized",
      "discriminator": [
        67,
        55,
        166,
        66,
        247,
        192,
        196,
        20
      ]
    },
    {
      "name": "factoryInitialized",
      "discriminator": [
        20,
        86,
        103,
        75,
        20,
        220,
        162,
        63
      ]
    }
  ],
  "errors": [
    {
      "code": 6000,
      "name": "notManager",
      "msg": "Signer is not the current factory manager"
    },
    {
      "code": 6001,
      "name": "notPendingManager",
      "msg": "Signer is not the pending manager"
    },
    {
      "code": 6002,
      "name": "factoryPaused",
      "msg": "Factory is paused"
    },
    {
      "code": 6003,
      "name": "factoryNotPaused",
      "msg": "Factory is not paused"
    },
    {
      "code": 6004,
      "name": "notOwner",
      "msg": "Signer is not the current asset class owner"
    },
    {
      "code": 6005,
      "name": "notPendingOwner",
      "msg": "Signer is not the pending asset class owner"
    },
    {
      "code": 6006,
      "name": "invalidVersion",
      "msg": "Version must be the asset class's latest version + 1"
    },
    {
      "code": 6007,
      "name": "versionNotDraft",
      "msg": "Asset class version is not in Draft state"
    },
    {
      "code": 6008,
      "name": "functionalityOutOfBounds",
      "msg": "Functionality is past the mask capacity"
    },
    {
      "code": 6009,
      "name": "overflow",
      "msg": "Arithmetic overflow"
    }
  ],
  "types": [
    {
      "name": "assetClassCreated",
      "type": {
        "kind": "struct",
        "fields": [
          {
            "name": "configId",
            "type": "u64"
          },
          {
            "name": "owner",
            "type": "pubkey"
          },
          {
            "name": "manager",
            "type": "pubkey"
          }
        ]
      }
    },
    {
      "name": "assetClassOwnership",
      "type": {
        "kind": "struct",
        "fields": [
          {
            "name": "owner",
            "type": "pubkey"
          },
          {
            "name": "latestVersion",
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
      "name": "assetClassPendingOwner",
      "type": {
        "kind": "struct",
        "fields": [
          {
            "name": "pendingOwner",
            "type": "pubkey"
          },
          {
            "name": "bump",
            "type": "u8"
          }
        ]
      }
    },
    {
      "name": "assetClassVersion",
      "serialization": "bytemuck",
      "repr": {
        "kind": "c"
      },
      "type": {
        "kind": "struct",
        "fields": [
          {
            "name": "configId",
            "type": "u64"
          },
          {
            "name": "version",
            "type": "u64"
          },
          {
            "name": "state",
            "type": "u8"
          },
          {
            "name": "bump",
            "type": "u8"
          },
          {
            "name": "padding",
            "type": {
              "array": [
                "u8",
                6
              ]
            }
          },
          {
            "name": "mask",
            "type": {
              "array": [
                "u8",
                1024
              ]
            }
          }
        ]
      }
    },
    {
      "name": "assetClassVersionFinalized",
      "type": {
        "kind": "struct",
        "fields": [
          {
            "name": "configId",
            "type": "u64"
          },
          {
            "name": "version",
            "type": "u64"
          },
          {
            "name": "owner",
            "type": "pubkey"
          }
        ]
      }
    },
    {
      "name": "assetClassVersionFunctionalitiesEnabled",
      "type": {
        "kind": "struct",
        "fields": [
          {
            "name": "configId",
            "type": "u64"
          },
          {
            "name": "version",
            "type": "u64"
          },
          {
            "name": "functionalities",
            "type": {
              "vec": "u16"
            }
          },
          {
            "name": "owner",
            "type": "pubkey"
          }
        ]
      }
    },
    {
      "name": "assetClassVersionInitialized",
      "type": {
        "kind": "struct",
        "fields": [
          {
            "name": "configId",
            "type": "u64"
          },
          {
            "name": "version",
            "type": "u64"
          },
          {
            "name": "owner",
            "type": "pubkey"
          }
        ]
      }
    },
    {
      "name": "factory",
      "type": {
        "kind": "struct",
        "fields": [
          {
            "name": "manager",
            "type": "pubkey"
          },
          {
            "name": "pause",
            "type": "bool"
          },
          {
            "name": "bump",
            "type": "u8"
          }
        ]
      }
    },
    {
      "name": "factoryInitialized",
      "type": {
        "kind": "struct",
        "fields": [
          {
            "name": "manager",
            "type": "pubkey"
          }
        ]
      }
    },
    {
      "name": "factoryPendingManager",
      "type": {
        "kind": "struct",
        "fields": [
          {
            "name": "pendingManager",
            "type": "pubkey"
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
