package ddd

type MessageTranslator func(from []byte) (Event, error)
