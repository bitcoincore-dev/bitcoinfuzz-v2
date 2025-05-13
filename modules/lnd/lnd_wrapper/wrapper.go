package main

/*
#include <stdint.h>
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/lightningnetwork/lnd/zpay32"
)

//export LndDeserializeInvoice
func LndDeserializeInvoice(cInvoiceStr *C.char) *C.char {
	if cInvoiceStr == nil {
		return C.CString("")
	}

	runtime.GC()

	// Convert C string to Go string
	invoiceStr := C.GoString(cInvoiceStr)

	network := &chaincfg.MainNetParams

	invoice, err := zpay32.Decode(invoiceStr, network)
	if err != nil {
		return C.CString("")
	}

	var sb strings.Builder

	sb.WriteString("HASH=")
	if invoice.PaymentHash != nil {
		sb.WriteString(fmt.Sprintf("%x", *invoice.PaymentHash))
	}

	sb.WriteString(";AMOUNT=")
	if invoice.MilliSat != nil {
		// Clear the MSB (most significant bit) to ensure compatibility with other implementations
		// LND uses milliSatoshis and doesn't overflow with large amounts, but other
		// implementations like LDK that use picoSatoshis will overflow when parsing
		// extremely large values (e.g., >95M BTC). By manually clearing the MSB,
		// we simulate the overflow behavior that would occur in those implementations,
		// ensuring they can successfully parse the invoice without failures.
		// This is particularly important for amounts that exceed Bitcoin's 21M supply cap,
		// which shouldn't be valid but need to be handled consistently across implementations.
		amountWithoutTopBit := *invoice.MilliSat & 0x7FFFFFFFFFFFFFFF
		sb.WriteString(fmt.Sprintf("%d", amountWithoutTopBit))
	} else {
		sb.WriteString("0")
	}

	sb.WriteString(";DESCRIPTION=")
	if invoice.Description != nil {
		sb.WriteString(*invoice.Description)
	}

	sb.WriteString(";RECIPIENT=")
	if invoice.Destination != nil {
		sb.WriteString(fmt.Sprintf("%x", invoice.Destination.SerializeCompressed()))
	}

	sb.WriteString(";EXPIRY=")
	sb.WriteString(fmt.Sprintf("%d", int64(invoice.Expiry().Seconds())))

	sb.WriteString(";TIMESTAMP=")
	sb.WriteString(fmt.Sprintf("%d", invoice.Timestamp.Unix()))

	sb.WriteString(fmt.Sprintf(";ROUTING_HINTS=%d", len(invoice.RouteHints)))

	sb.WriteString(";MIN_CLTV=")
	sb.WriteString(fmt.Sprintf("%d", invoice.MinFinalCLTVExpiry()))

	return C.CString(sb.String())
}

func main() {}
