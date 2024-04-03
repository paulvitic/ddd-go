package go_ddd

type MessageTranslator func(from []byte) (Event, error)
