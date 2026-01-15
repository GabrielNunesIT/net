// Package netconferrors provides RFC 6241 NETCONF error constants and pre-built error values.
package netconferrors

import "github.com/damianoneill/net/v2/netconf/common"

// Error severity values as defined in RFC 6241 (unexported).
const (
	severityError   = "error"
	severityWarning = "warning"
)

// Error type values as defined in RFC 6241 (unexported).
const (
	errorTypeTransport   = "transport"
	errorTypeRPC         = "rpc"
	errorTypeProtocol    = "protocol"
	errorTypeApplication = "application"
)

// Error tag values as defined in RFC 6241 Appendix A (unexported).
const (
	errTagInUse               = "in-use"
	errTagInvalidValue        = "invalid-value"
	errTagTooBig              = "too-big"
	errTagMissingAttribute    = "missing-attribute"
	errTagBadAttribute        = "bad-attribute"
	errTagUnknownAttribute    = "unknown-attribute"
	errTagMissingElement      = "missing-element"
	errTagBadElement          = "bad-element"
	errTagUnknownElement      = "unknown-element"
	errTagUnknownNamespace    = "unknown-namespace"
	errTagAccessDenied        = "access-denied"
	errTagLockDenied          = "lock-denied"
	errTagResourceDenied      = "resource-denied"
	errTagRollbackFailed      = "rollback-failed"
	errTagDataExists          = "data-exists"
	errTagDataMissing         = "data-missing"
	errTagOperationNotSupported = "operation-not-supported"
	errTagOperationFailed     = "operation-failed"
	errTagMalformedMessage    = "malformed-message"
)

// Pre-built RPCError values for common error conditions.
// Use these as templates; copy and set the Message field as needed.
var (
	// ErrInUse indicates the request requires a resource that is already in use.
	ErrInUse = common.RPCError{Type: errorTypeProtocol, Tag: errTagInUse, Severity: severityError}

	// ErrInvalidValue indicates the request contains an unacceptable value.
	ErrInvalidValue = common.RPCError{Type: errorTypeProtocol, Tag: errTagInvalidValue, Severity: severityError}

	// ErrTooBig indicates the response content exceeds the maximum size.
	ErrTooBig = common.RPCError{Type: errorTypeTransport, Tag: errTagTooBig, Severity: severityError}

	// ErrMissingAttribute indicates a required attribute is missing.
	ErrMissingAttribute = common.RPCError{Type: errorTypeRPC, Tag: errTagMissingAttribute, Severity: severityError}

	// ErrBadAttribute indicates an attribute value is not valid.
	ErrBadAttribute = common.RPCError{Type: errorTypeRPC, Tag: errTagBadAttribute, Severity: severityError}

	// ErrUnknownAttribute indicates an unrecognized attribute was present.
	ErrUnknownAttribute = common.RPCError{Type: errorTypeRPC, Tag: errTagUnknownAttribute, Severity: severityError}

	// ErrMissingElement indicates a required element is missing.
	ErrMissingElement = common.RPCError{Type: errorTypeProtocol, Tag: errTagMissingElement, Severity: severityError}

	// ErrBadElement indicates an element value is not valid.
	ErrBadElement = common.RPCError{Type: errorTypeProtocol, Tag: errTagBadElement, Severity: severityError}

	// ErrUnknownElement indicates an unrecognized element was present.
	ErrUnknownElement = common.RPCError{Type: errorTypeProtocol, Tag: errTagUnknownElement, Severity: severityError}

	// ErrUnknownNamespace indicates an unexpected namespace was present.
	ErrUnknownNamespace = common.RPCError{Type: errorTypeProtocol, Tag: errTagUnknownNamespace, Severity: severityError}

	// ErrAccessDenied indicates access to the requested resource was denied.
	ErrAccessDenied = common.RPCError{Type: errorTypeProtocol, Tag: errTagAccessDenied, Severity: severityError}

	// ErrLockDenied indicates the lock is held by another entity.
	ErrLockDenied = common.RPCError{Type: errorTypeProtocol, Tag: errTagLockDenied, Severity: severityError}

	// ErrResourceDenied indicates insufficient resources to complete the request.
	ErrResourceDenied = common.RPCError{Type: errorTypeProtocol, Tag: errTagResourceDenied, Severity: severityError}

	// ErrRollbackFailed indicates the rollback operation failed.
	ErrRollbackFailed = common.RPCError{Type: errorTypeProtocol, Tag: errTagRollbackFailed, Severity: severityError}

	// ErrDataExists indicates the data already exists (for create operations).
	ErrDataExists = common.RPCError{Type: errorTypeApplication, Tag: errTagDataExists, Severity: severityError}

	// ErrDataMissing indicates required data is missing.
	ErrDataMissing = common.RPCError{Type: errorTypeApplication, Tag: errTagDataMissing, Severity: severityError}

	// ErrOperationNotSupported indicates the requested operation is not supported.
	ErrOperationNotSupported = common.RPCError{Type: errorTypeProtocol, Tag: errTagOperationNotSupported, Severity: severityError}

	// ErrOperationFailed indicates the operation failed for an unspecified reason.
	ErrOperationFailed = common.RPCError{Type: errorTypeProtocol, Tag: errTagOperationFailed, Severity: severityError}

	// ErrMalformedMessage indicates malformed XML was received (base:1.1 only).
	ErrMalformedMessage = common.RPCError{Type: errorTypeRPC, Tag: errTagMalformedMessage, Severity: severityError}
)

// WithMessage returns a copy of the error with the specified message.
func WithMessage(err common.RPCError, message string) common.RPCError {
	err.Message = message
	return err
}

// WithPath returns a copy of the error with the specified error-path.
func WithPath(err common.RPCError, path string) common.RPCError {
	err.Path = path
	return err
}
