package main

import (
	"log"
	"strconv"
	"strings"

	"github.com/vugu/vugu"
)

type DistanceInputComponent struct {
	BindValue *Distance
	BindLock  *bool

	AttrMap vugu.AttrMap
}

func (c *DistanceInputComponent) handleValueChange(event vugu.DOMEvent) {
	strVal := event.PropString("target", "value")
	strVal = strings.ReplaceAll(strVal, ",", ".")

	val, err := strconv.ParseFloat(strVal, 64)
	if err != nil {
		log.Printf("strconv.ParseFloat() failed: %v", err)
		return
	}

	*c.BindValue = Distance(val)
}

func (c *DistanceInputComponent) handleLockChange(event vugu.DOMEvent) {
	*c.BindLock = event.PropBool("target", "checked")
}
