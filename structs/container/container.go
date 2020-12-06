package container

type Container interface {
	// Len return the number of elements in container.
	Len() int
	// Insert insert into specified element, return false if duplicate.
	Insert(element Comparer) bool
	// Delete delete and return specified element, return nil if not found.
	Delete(element Comparer) Comparer
	// Search search the specified element, return nil if not found.
	Search(element Comparer) Comparer
	// Iter return a Iterator, include element: start <= k <= boundary.
	Iter(start Comparer, boundary Comparer) Iterator
}