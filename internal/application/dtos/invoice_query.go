// Package dtos re-exports persistence query types for transitional legacy code.
// New code must import from pkg/persistence directly.
package dtos

import "github.com/jailtonjunior94/financialcontrol-api/pkg/persistence"

type InvoiceQuery = persistence.InvoiceQuery
type InvoiceRead = persistence.InvoiceRead
type InvoiceItemRead = persistence.InvoiceItemRead
type CategoryRead = persistence.CategoryRead
type BillQuery = persistence.BillQuery
type BillItemQuery = persistence.BillItemQuery
