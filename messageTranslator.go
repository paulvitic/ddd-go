package go_ddd

type EventTranslator func(from []byte) (Event, error)
