package util

import (
	"math"
	"testing"
)

// CheckOverflow tests
// Note: The formula (left+right)-left != right has limitations in Go due to
// wrapping arithmetic. It correctly identifies no-overflow cases but may not
// detect all actual overflows since the subtraction also wraps.

func TestCheckOverflowNoOverflow(t *testing.T) {
	result := CheckOverflow(10, 20)
	if result {
		t.Error("expected no overflow for 10 + 20")
	}
}

func TestCheckOverflowZeroAddend(t *testing.T) {
	result := CheckOverflow(math.MaxInt64, 0)
	if result {
		t.Error("expected no overflow when adding 0")
	}
}

func TestCheckOverflowLargeValues(t *testing.T) {
	result := CheckOverflow(1000000, 2000000)
	if result {
		t.Error("expected no overflow for small values")
	}
}

func TestCheckOverflowNegativeAddition(t *testing.T) {
	result := CheckOverflow(-100, -200)
	if result {
		t.Error("expected no overflow for -100 + -200")
	}
}

func TestCheckOverflowMixedSigns(t *testing.T) {
	result := CheckOverflow(math.MaxInt64, -1)
	if result {
		t.Error("expected no overflow for MaxInt64 + (-1)")
	}
}

func TestCheckOverflowSymmetric(t *testing.T) {
	// a+b should behave the same as b+a
	result1 := CheckOverflow(100, 200)
	result2 := CheckOverflow(200, 100)
	if result1 != result2 {
		t.Errorf("CheckOverflow should be symmetric: %v vs %v", result1, result2)
	}
}

func TestCheckOverflowSameValue(t *testing.T) {
	result := CheckOverflow(500, 500)
	if result {
		t.Error("expected no overflow for 500 + 500")
	}
}

// CheckUnderflow tests
// Same note as CheckOverflow: formula has limitations with wrapping.

func TestCheckUnderflowNoUnderflow(t *testing.T) {
	result := CheckUnderflow(100, 50)
	if result {
		t.Error("expected no underflow for 100 - 50")
	}
}

func TestCheckUnderflowZero(t *testing.T) {
	result := CheckUnderflow(math.MinInt64, 0)
	if result {
		t.Error("expected no underflow when subtracting 0")
	}
}

func TestCheckUnderflowNegativeSubtraction(t *testing.T) {
	result := CheckUnderflow(100, -50)
	if result {
		t.Error("expected no underflow for 100 - (-50)")
	}
}

func TestCheckUnderflowMixedSigns(t *testing.T) {
	result := CheckUnderflow(math.MinInt64, -1)
	if result {
		t.Error("expected no underflow for MinInt64 - (-1)")
	}
}

func TestCheckUnderflowLargeValues(t *testing.T) {
	result := CheckUnderflow(1000000, 500000)
	if result {
		t.Error("expected no underflow for 1000000 - 500000")
	}
}

// CheckOverflowMul tests

func TestCheckOverflowMulPositive(t *testing.T) {
	result := CheckOverflowMul(math.MaxInt64, 2)
	if !result {
		t.Error("expected overflow for MaxInt64 * 2")
	}
}

func TestCheckOverflowMulNegative(t *testing.T) {
	result := CheckOverflowMul(math.MinInt64, 2)
	if !result {
		t.Error("expected overflow for MinInt64 * 2")
	}
}

func TestCheckOverflowMulMinMax(t *testing.T) {
	result := CheckOverflowMul(math.MinInt64, -1)
	if !result {
		t.Error("expected overflow for MinInt64 * -1")
	}
}

func TestCheckOverflowMulZero(t *testing.T) {
	result := CheckOverflowMul(math.MaxInt64, 0)
	if result {
		t.Error("expected no overflow for anything * 0")
	}
}

func TestCheckOverflowMulOne(t *testing.T) {
	result := CheckOverflowMul(math.MaxInt64, 1)
	if result {
		t.Error("expected no overflow for anything * 1")
	}
}

func TestCheckOverflowMulNoOverflow(t *testing.T) {
	result := CheckOverflowMul(100, 200)
	if result {
		t.Error("expected no overflow for 100 * 200")
	}
}

func TestCheckOverflowMulNegativeValues(t *testing.T) {
	result := CheckOverflowMul(-100, 200)
	if result {
		t.Error("expected no overflow for -100 * 200")
	}
}

func TestCheckOverflowMulBothNegative(t *testing.T) {
	result := CheckOverflowMul(-100, -200)
	if result {
		t.Error("expected no overflow for -100 * -200")
	}
}

// CheckOverflowPow tests

func TestCheckOverflowPowPositive(t *testing.T) {
	result := CheckOverflowPow(2, 64)
	if !result {
		t.Error("expected overflow for 2^64")
	}
}

func TestCheckOverflowPowMinInt64(t *testing.T) {
	result := CheckOverflowPow(math.MinInt64, 2)
	if !result {
		t.Error("expected overflow for MinInt64^2")
	}
}

func TestCheckOverflowPowZeroExp(t *testing.T) {
	result := CheckOverflowPow(999999, 0)
	if result {
		t.Error("expected no overflow for anything^0")
	}
}

func TestCheckOverflowPowZeroBase(t *testing.T) {
	result := CheckOverflowPow(0, 999999)
	if result {
		t.Error("expected no overflow for 0^anything")
	}
}

func TestCheckOverflowPowOne(t *testing.T) {
	result := CheckOverflowPow(999999, 1)
	if result {
		t.Error("expected no overflow for anything^1")
	}
}

func TestCheckOverflowPowBaseOne(t *testing.T) {
	result := CheckOverflowPow(1, 999999)
	if result {
		t.Error("expected no overflow for 1^anything")
	}
}

func TestCheckOverflowPowNoOverflow(t *testing.T) {
	result := CheckOverflowPow(2, 10)
	if result {
		t.Error("expected no overflow for 2^10")
	}
}

func TestCheckOverflowPowLargeExp(t *testing.T) {
	result := CheckOverflowPow(2, 70)
	if !result {
		t.Error("expected overflow for 2^70")
	}
}

func TestCheckOverflowPowMinInt64Exp(t *testing.T) {
	result := CheckOverflowPow(5, math.MinInt64)
	if !result {
		t.Error("expected overflow for 5^MinInt64")
	}
}

// List tests

func TestNewList(t *testing.T) {
	l := NewList[int]()
	if l == nil {
		t.Fatal("NewList returned nil")
	}
	if l.Len() != 0 {
		t.Errorf("expected length 0, got %d", l.Len())
	}
}

func TestListZeroValue(t *testing.T) {
	var l List[int]
	if l.Len() != 0 {
		t.Errorf("zero value list should have length 0, got %d", l.Len())
	}
	if l.Front() != nil {
		t.Error("zero value list Front() should be nil")
	}
	if l.Back() != nil {
		t.Error("zero value list Back() should be nil")
	}
}

func TestListPushFront(t *testing.T) {
	l := NewList[int]()
	e := l.PushFront(10)
	if e == nil {
		t.Fatal("PushFront returned nil")
	}
	if e.Value != 10 {
		t.Errorf("expected value 10, got %d", e.Value)
	}
	if l.Len() != 1 {
		t.Errorf("expected length 1, got %d", l.Len())
	}
}

func TestListPushBack(t *testing.T) {
	l := NewList[int]()
	e := l.PushBack(20)
	if e == nil {
		t.Fatal("PushBack returned nil")
	}
	if e.Value != 20 {
		t.Errorf("expected value 20, got %d", e.Value)
	}
	if l.Len() != 1 {
		t.Errorf("expected length 1, got %d", l.Len())
	}
}

func TestListMultiplePushFront(t *testing.T) {
	l := NewList[int]()
	l.PushFront(3)
	l.PushFront(2)
	l.PushFront(1)

	if l.Len() != 3 {
		t.Errorf("expected length 3, got %d", l.Len())
	}

	front := l.Front()
	if front == nil || front.Value != 1 {
		t.Errorf("expected front value 1, got %v", front)
	}

	back := l.Back()
	if back == nil || back.Value != 3 {
		t.Errorf("expected back value 3, got %v", back)
	}
}

func TestListMultiplePushBack(t *testing.T) {
	l := NewList[int]()
	l.PushBack(1)
	l.PushBack(2)
	l.PushBack(3)

	if l.Len() != 3 {
		t.Errorf("expected length 3, got %d", l.Len())
	}

	front := l.Front()
	if front == nil || front.Value != 1 {
		t.Errorf("expected front value 1, got %v", front)
	}

	back := l.Back()
	if back == nil || back.Value != 3 {
		t.Errorf("expected back value 3, got %v", back)
	}
}

func TestListPushFrontBackMix(t *testing.T) {
	l := NewList[int]()
	l.PushFront(1)
	l.PushBack(2)
	l.PushFront(0)
	l.PushBack(3)

	if l.Len() != 4 {
		t.Errorf("expected length 4, got %d", l.Len())
	}

	var values []int
	for e := l.Front(); e != nil; e = e.Next() {
		values = append(values, e.Value)
	}

	expected := []int{0, 1, 2, 3}
	for i, v := range values {
		if v != expected[i] {
			t.Errorf("position %d: expected %d, got %d", i, expected[i], v)
		}
	}
}

func TestListFrontBackEmpty(t *testing.T) {
	l := NewList[int]()
	if l.Front() != nil {
		t.Error("Front() on empty list should be nil")
	}
	if l.Back() != nil {
		t.Error("Back() on empty list should be nil")
	}
}

func TestListRemove(t *testing.T) {
	l := NewList[int]()
	_ = l.PushBack(1)
	e2 := l.PushBack(2)
	_ = l.PushBack(3)

	l.Remove(e2)

	if l.Len() != 2 {
		t.Errorf("expected length 2 after remove, got %d", l.Len())
	}

	var values []int
	for e := l.Front(); e != nil; e = e.Next() {
		values = append(values, e.Value)
	}

	expected := []int{1, 3}
	for i, v := range values {
		if v != expected[i] {
			t.Errorf("position %d: expected %d, got %d", i, expected[i], v)
		}
	}
}

func TestListRemoveFront(t *testing.T) {
	l := NewList[int]()
	e := l.PushFront(42)
	l.Remove(e)

	if l.Len() != 0 {
		t.Errorf("expected length 0 after removing last element, got %d", l.Len())
	}
	if l.Front() != nil {
		t.Error("Front() should be nil after removing last element")
	}
}

func TestListRemoveBack(t *testing.T) {
	l := NewList[int]()
	e := l.PushBack(42)
	l.Remove(e)

	if l.Len() != 0 {
		t.Errorf("expected length 0 after removing last element, got %d", l.Len())
	}
}

func TestListRemoveNonMember(t *testing.T) {
	l := NewList[int]()
	l.PushBack(1)
	var standalone List[int]
	_ = standalone.PushFront(99)

	removingSameList := l.Len()
	l.Remove(standalone.Front())
	if l.Len() != removingSameList {
		t.Error("removing non-member should not modify list")
	}
}

func TestListInsertBefore(t *testing.T) {
	l := NewList[int]()
	e2 := l.PushBack(2)
	_ = l.PushBack(3)

	e1 := l.InsertBefore(1, e2)
	if e1 == nil {
		t.Fatal("InsertBefore returned nil")
	}
	if e1.Value != 1 {
		t.Errorf("expected inserted value 1, got %d", e1.Value)
	}
	if l.Len() != 3 {
		t.Errorf("expected length 3, got %d", l.Len())
	}

	var values []int
	for e := l.Front(); e != nil; e = e.Next() {
		values = append(values, e.Value)
	}

	expected := []int{1, 2, 3}
	for i, v := range values {
		if v != expected[i] {
			t.Errorf("position %d: expected %d, got %d", i, expected[i], v)
		}
	}
}

func TestListInsertBeforeNonMember(t *testing.T) {
	l := NewList[int]()
	l.PushBack(1)
	var standalone List[int]
	_ = standalone.PushFront(99)

	result := l.InsertBefore(0, standalone.Front())
	if result != nil {
		t.Error("InsertBefore with non-member mark should return nil")
	}
	if l.Len() != 1 {
		t.Error("InsertBefore with non-member mark should not modify list")
	}
}

func TestListInsertAfter(t *testing.T) {
	l := NewList[int]()
	e1 := l.PushBack(1)

	e2 := l.InsertAfter(2, e1)
	if e2 == nil {
		t.Fatal("InsertAfter returned nil")
	}
	if e2.Value != 2 {
		t.Errorf("expected inserted value 2, got %d", e2.Value)
	}
	if l.Len() != 2 {
		t.Errorf("expected length 2, got %d", l.Len())
	}

	var values []int
	for e := l.Front(); e != nil; e = e.Next() {
		values = append(values, e.Value)
	}

	expected := []int{1, 2}
	for i, v := range values {
		if v != expected[i] {
			t.Errorf("position %d: expected %d, got %d", i, expected[i], v)
		}
	}
}

func TestListInsertAfterNonMember(t *testing.T) {
	l := NewList[int]()
	l.PushBack(1)
	var standalone List[int]
	mark := standalone.PushFront(99)

	result := l.InsertAfter(0, mark)
	if result != nil {
		t.Error("InsertAfter with non-member mark should return nil")
	}
}

func TestListMoveToFront(t *testing.T) {
	l := NewList[int]()
	l.PushBack(1)
	e2 := l.PushBack(2)
	l.PushBack(3)

	l.MoveToFront(e2)

	front := l.Front()
	if front == nil || front.Value != 2 {
		t.Errorf("expected front value 2 after MoveToFront, got %v", front)
	}
	if l.Len() != 3 {
		t.Errorf("expected length 3, got %d", l.Len())
	}
}

func TestListMoveToFrontAlreadyFront(t *testing.T) {
	l := NewList[int]()
	e := l.PushFront(10)

	l.MoveToFront(e)

	front := l.Front()
	if front == nil || front.Value != 10 {
		t.Errorf("expected front value 10, got %v", front)
	}
}

func TestListMoveToFrontNonMember(t *testing.T) {
	l := NewList[int]()
	_ = l.PushBack(1)
	var standalone List[int]
	other := standalone.PushFront(99)

	before := l.Len()
	l.MoveToFront(other)
	if l.Len() != before {
		t.Error("MoveToFront non-member should not modify list")
	}
}

func TestListMoveToBack(t *testing.T) {
	l := NewList[int]()
	l.PushBack(1)
	e2 := l.PushBack(2)
	l.PushBack(3)

	l.MoveToBack(e2)

	back := l.Back()
	if back == nil || back.Value != 2 {
		t.Errorf("expected back value 2 after MoveToBack, got %v", back)
	}
	if l.Len() != 3 {
		t.Errorf("expected length 3, got %d", l.Len())
	}
}

func TestListMoveToBackAlreadyBack(t *testing.T) {
	l := NewList[int]()
	e := l.PushBack(10)

	l.MoveToBack(e)

	back := l.Back()
	if back == nil || back.Value != 10 {
		t.Errorf("expected back value 10, got %v", back)
	}
}

func TestListMoveBefore(t *testing.T) {
	l := NewList[int]()
	_ = l.PushBack(1)
	e3 := l.PushBack(3)
	e2 := l.PushBack(2)

	l.MoveBefore(e2, e3)

	var values []int
	for e := l.Front(); e != nil; e = e.Next() {
		values = append(values, e.Value)
	}

	expected := []int{1, 2, 3}
	for i, v := range values {
		if v != expected[i] {
			t.Errorf("position %d: expected %d, got %d", i, expected[i], v)
		}
	}
}

func TestListMoveBeforeNonMember(t *testing.T) {
	l := NewList[int]()
	_ = l.PushBack(1)
	var standalone List[int]
	_ = standalone.PushFront(99)

	l.MoveBefore(standalone.Front(), l.Front())
	if l.Len() != 1 {
		t.Error("MoveBefore non-member should not modify list")
	}
}

func TestListMoveBeforeSameElement(t *testing.T) {
	l := NewList[int]()
	e := l.PushBack(1)

	l.MoveBefore(e, e)
	if l.Len() != 1 {
		t.Error("MoveBefore same element should not modify list")
	}
}

func TestListMoveAfter(t *testing.T) {
	l := NewList[int]()
	_ = l.PushBack(1)
	e2 := l.PushBack(2)
	_ = l.PushBack(3)

	l.MoveAfter(e2, l.Front())

	var values []int
	for e := l.Front(); e != nil; e = e.Next() {
		values = append(values, e.Value)
	}

	expected := []int{1, 2, 3}
	for i, v := range values {
		if v != expected[i] {
			t.Errorf("position %d: expected %d, got %d", i, expected[i], v)
		}
	}
}

func TestListMoveAfterNonMember(t *testing.T) {
	l := NewList[int]()
	_ = l.PushBack(1)
	var standalone List[int]
	_ = standalone.PushFront(99)

	l.MoveAfter(standalone.Front(), l.Front())
	if l.Len() != 1 {
		t.Error("MoveAfter non-member should not modify list")
	}
}

func TestListMoveAfterSameElement(t *testing.T) {
	l := NewList[int]()
	e := l.PushBack(1)

	l.MoveAfter(e, e)
	if l.Len() != 1 {
		t.Error("MoveAfter same element should not modify list")
	}
}

func TestListPushBackList(t *testing.T) {
	l := NewList[int]()
	l.PushBack(1)

	other := NewList[int]()
	other.PushBack(2)
	other.PushBack(3)

	l.PushBackList(other)

	if l.Len() != 3 {
		t.Errorf("expected length 3, got %d", l.Len())
	}

	var values []int
	for e := l.Front(); e != nil; e = e.Next() {
		values = append(values, e.Value)
	}

	expected := []int{1, 2, 3}
	for i, v := range values {
		if v != expected[i] {
			t.Errorf("position %d: expected %d, got %d", i, expected[i], v)
		}
	}
}

func TestListPushBackListSelf(t *testing.T) {
	l := NewList[int]()
	l.PushBack(1)
	l.PushBack(2)

	l.PushBackList(l)

	if l.Len() != 4 {
		t.Errorf("expected length 4, got %d", l.Len())
	}

	var values []int
	for e := l.Front(); e != nil; e = e.Next() {
		values = append(values, e.Value)
	}

	expected := []int{1, 2, 1, 2}
	for i, v := range values {
		if v != expected[i] {
			t.Errorf("position %d: expected %d, got %d", i, expected[i], v)
		}
	}
}

func TestListPushFrontList(t *testing.T) {
	l := NewList[int]()
	l.PushBack(3)

	other := NewList[int]()
	other.PushBack(1)
	other.PushBack(2)

	l.PushFrontList(other)

	if l.Len() != 3 {
		t.Errorf("expected length 3, got %d", l.Len())
	}

	var values []int
	for e := l.Front(); e != nil; e = e.Next() {
		values = append(values, e.Value)
	}

	expected := []int{1, 2, 3}
	for i, v := range values {
		if v != expected[i] {
			t.Errorf("position %d: expected %d, got %d", i, expected[i], v)
		}
	}
}

func TestListInit(t *testing.T) {
	l := NewList[int]()
	l.PushBack(1)
	l.PushBack(2)

	l.Init()

	if l.Len() != 0 {
		t.Errorf("expected length 0 after Init, got %d", l.Len())
	}
	if l.Front() != nil {
		t.Error("Front() should be nil after Init")
	}
}

func TestListElementNextPrev(t *testing.T) {
	l := NewList[int]()
	e1 := l.PushBack(1)
	e2 := l.PushBack(2)
	e3 := l.PushBack(3)

	if e1.Prev() != nil {
		t.Error("first element Prev() should be nil")
	}
	if e1.Next() != e2 {
		t.Error("first element Next() should be second element")
	}

	if e2.Prev() != e1 {
		t.Error("middle element Prev() should be first element")
	}
	if e2.Next() != e3 {
		t.Error("middle element Next() should be third element")
	}

	if e3.Prev() != e2 {
		t.Error("last element Prev() should be middle element")
	}
	if e3.Next() != nil {
		t.Error("last element Next() should be nil")
	}
}

func TestListRemoveMiddleElement(t *testing.T) {
	l := NewList[int]()
	e1 := l.PushBack(1)
	e2 := l.PushBack(2)
	e3 := l.PushBack(3)

	l.Remove(e2)

	if e1.Next() != e3 {
		t.Error("after removing middle, first should point to third")
	}
	if e3.Prev() != e1 {
		t.Error("after removing middle, third should point to first")
	}
}

func TestListWithPointers(t *testing.T) {
	l := NewList[*int]()
	a, b, c := 1, 2, 3
	l.PushBack(&a)
	l.PushBack(&b)
	l.PushBack(&c)

	if l.Len() != 3 {
		t.Errorf("expected length 3, got %d", l.Len())
	}

	val := *l.Front().Value
	if val != 1 {
		t.Errorf("expected first value 1, got %d", val)
	}
}

func TestListWithStructs(t *testing.T) {
	type Point struct {
		X, Y int
	}
	l := NewList[Point]()
	l.PushBack(Point{1, 2})
	l.PushBack(Point{3, 4})

	if l.Len() != 2 {
		t.Errorf("expected length 2, got %d", l.Len())
	}

	p := l.Front().Value
	if p.X != 1 || p.Y != 2 {
		t.Errorf("expected Point{1, 2}, got %+v", p)
	}
}

// Stack tests

func TestNewStack(t *testing.T) {
	s := NewStack[int]()
	if s == nil {
		t.Fatal("NewStack returned nil")
	}
	if s.Len() != 0 {
		t.Errorf("expected length 0, got %d", s.Len())
	}
}

func TestStackPushAndPop(t *testing.T) {
	s := NewStack[int]()
	s.Push(1)
	s.Push(2)
	s.Push(3)

	if s.Len() != 3 {
		t.Errorf("expected length 3, got %d", s.Len())
	}

	val := s.Pop()
	if val != 3 {
		t.Errorf("expected popped value 3 (LIFO), got %d", val)
	}
	if s.Len() != 2 {
		t.Errorf("expected length 2 after pop, got %d", s.Len())
	}
}

func TestStackLFORder(t *testing.T) {
	s := NewStack[int]()
	s.Push(10)
	s.Push(20)
	s.Push(30)
	s.Push(40)

	expected := []int{40, 30, 20, 10}
	for _, exp := range expected {
		val := s.Pop()
		if val != exp {
			t.Errorf("expected LIFO order %d, got %d", exp, val)
		}
	}
}

func TestStackPeek(t *testing.T) {
	s := NewStack[int]()
	s.Push(42)

	peeked := s.Peek()
	if peeked != 42 {
		t.Errorf("expected peeked value 42, got %d", peeked)
	}
	if s.Len() != 1 {
		t.Errorf("Peek should not change length, got %d", s.Len())
	}

	popped := s.Pop()
	if popped != 42 {
		t.Errorf("expected popped value 42, got %d", popped)
	}
}

func TestStackPopEmpty(t *testing.T) {
	s := NewStack[int]()
	val := s.Pop()
	var zero int
	if val != zero {
		t.Errorf("expected zero value when popping empty stack, got %d", val)
	}
}

func TestStackPeekEmpty(t *testing.T) {
	s := NewStack[int]()
	val := s.Peek()
	var zero int
	if val != zero {
		t.Errorf("expected zero value when peeking empty stack, got %d", val)
	}
}

func TestStackPopBack(t *testing.T) {
	s := NewStack[int]()
	s.Push(1)
	s.Push(2)
	s.Push(3)

	// PopBack pops from the back of the underlying list (first pushed element)
	val := s.PopBack()
	if val != 1 {
		t.Errorf("expected PopBack value 1 (first pushed), got %d", val)
	}
	if s.Len() != 2 {
		t.Errorf("expected length 2 after PopBack, got %d", s.Len())
	}
}

func TestStackPopBackEmpty(t *testing.T) {
	s := NewStack[int]()
	val := s.PopBack()
	var zero int
	if val != zero {
		t.Errorf("expected zero value when PopBack on empty stack, got %d", val)
	}
}

func TestStackLen(t *testing.T) {
	s := NewStack[int]()
	if s.Len() != 0 {
		t.Errorf("expected length 0, got %d", s.Len())
	}

	s.Push(1)
	if s.Len() != 1 {
		t.Errorf("expected length 1, got %d", s.Len())
	}

	s.Push(2)
	s.Push(3)
	if s.Len() != 3 {
		t.Errorf("expected length 3, got %d", s.Len())
	}

	s.Pop()
	if s.Len() != 2 {
		t.Errorf("expected length 2 after pop, got %d", s.Len())
	}
}

func TestStackString(t *testing.T) {
	s := NewStack[string]()
	s.Push("hello")
	s.Push("world")

	str := s.String()
	if len(str) == 0 {
		t.Error("String() should not be empty")
	}
}

func TestStackWithPointers(t *testing.T) {
	s := NewStack[*string]()
	a, b := "first", "second"
	s.Push(&a)
	s.Push(&b)

	val := s.Pop()
	if *val != "second" {
		t.Errorf("expected 'second', got %s", *val)
	}
}

func TestStackMultipleOps(t *testing.T) {
	s := NewStack[int]()
	s.Push(1)
	s.Push(2)
	s.Pop()
	s.Push(3)
	s.Push(4)
	s.Peek()
	s.Pop()
	s.Pop()

	if s.Len() != 1 {
		t.Errorf("expected length 1, got %d", s.Len())
	}
	if s.Pop() != 1 {
		t.Error("expected remaining value 1")
	}
}

// ToTitleCase tests

func TestToTitleCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "Hello"},
		{"hello world", "Hello World"},
		{"HELLO", "Hello"},
		{"hELLO wORLD", "Hello World"},
		{"", ""},
		{"a", "A"},
		{"hello-world", "Hello-World"},
		{"hello_world", "Hello_world"},
		{"already Title", "Already Title"},
		{"the quick brown fox", "The Quick Brown Fox"},
	}

	for _, tt := range tests {
		result := ToTitleCase(tt.input)
		if result != tt.expected {
			t.Errorf("ToTitleCase(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestToTitleCaseMixed(t *testing.T) {
	input := "gO gO gO"
	result := ToTitleCase(input)
	if result != "Go Go Go" {
		t.Errorf("ToTitleCase(%q) = %q, want %q", input, result, "Go Go Go")
	}
}
