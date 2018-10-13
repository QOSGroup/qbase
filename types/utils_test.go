package types

import (
	"bytes"
	"fmt"
	"testing"
)

func TestBool2Byte(t *testing.T) {
	bytrue := Bool2Byte(true)
	byfalse := Bool2Byte(false)

	if bytrue == nil || byfalse == nil {
		t.Error("Bool2Byte error")
		return
	}

	if bytes.Compare(bytrue, byfalse) > 0 {
		fmt.Print("Bool2Byte OK")
		return
	}
	t.Error("Bool2Byte error")
}

func TestInt2Byte(t *testing.T) {
	bysmall := Int2Byte(23)
	bylarge := Int2Byte(2345678910)

	if bysmall == nil || bylarge == nil {
		t.Error("Int2Byte error")
		return
	}
	fmt.Printf("small: %s, large: %s", string(bysmall), string(bylarge))
}

func TestCheckQsc(t *testing.T) {
	strOk := "qos34"
	strErr := "qsc$df3"
	strErr1 := "qsc34567891"

	if CheckQsc(strOk) && !CheckQsc(strErr) && !CheckQsc(strErr1) {
		fmt.Print("CheckQsc right")
		return
	}
	fmt.Print("ChechQsc error")
}
