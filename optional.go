package gpt3encoder

type Optional[T any] struct {
	data *T
}

func EmptyOption[T any]() Optional[T] {
	return Optional[T]{nil}
}

func OfOption[T any](data *T) Optional[T] {
	return Optional[T]{data}
}

func (o Optional[T]) IsPresent() bool {
	return o.data != nil
}

func (o Optional[T]) Get() T {
	return *o.data
}

func (o Optional[T]) GetOr(other T) T {
	if o.data == nil {
		return other
	}
	return *o.data
}
