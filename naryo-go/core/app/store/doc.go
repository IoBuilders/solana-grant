// Package store defines the application ports used to persist chain data
// ingested from a Node.
//
// A Store is chosen by matching a Node's store.ActiveConfiguration against
// Supports, and then receives only the data to persist.
package store
