// Package decoder defines the port through which raw on-chain payload bytes
// are decoded into ContractEventParameter values, per the filter.Strategy
// that produced them. Concrete implementations (e.g. an Anchor/Borsh decoder)
// live in infrastructure, since they depend on third-party chain SDKs that
// must never leak into the domain.
package decoder
