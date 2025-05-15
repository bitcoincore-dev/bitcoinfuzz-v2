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
		// Simulate uint64 overflow behavior for compatibility with all implementations.
		// LND uses milliSatoshis (10^-11 BTC) and does not overflow when handling very large values.
		// However, other implementations like LDK, which use picoBTC (10^-12 BTC), may overflow when parsing
		// extremely large amounts (e.g., >95M BTC). To simulate this behavior, we multiply and divide by 10.
		// This cancels out the value mathematically but mimics the overflow behavior that would occur in those
		// implementations. This is especially relevant for values exceeding Bitcoin’s 21M BTC cap—
		// which should be considered invalid—but must be handled in a consistent way across implementations
		// to avoid parsing errors.
		amountOverflow := (*invoice.MilliSat * 10) / 10
		sb.WriteString(fmt.Sprintf("%d", amountOverflow))
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
