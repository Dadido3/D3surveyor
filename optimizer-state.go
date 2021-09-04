// Copyright (C) 2021 David Vogel
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"log"
	"sync"
	"time"

	"github.com/vugu/vugu"
)

// OptimizerState stores the state of the optimizer and handles start/stop queries.
type OptimizerState struct {
	sync.RWMutex

	site     *Site
	running  bool // Optimizer is running.
	stopFlag bool // Flag signalling that the optimizer should stop.
}

// Running returns whether the optimizer is running or not.
func (os *OptimizerState) Running() bool {
	os.RLock()
	defer os.RUnlock()

	return os.running
}

func (os *OptimizerState) Start(event vugu.DOMEvent) {
	os.Lock()
	defer os.Unlock()

	// Check if there is already an optimizer running for this site.
	if os.running {
		return
	}
	os.running, os.stopFlag = true, false

	// Create clone of site to optimize. Also create a pair of tweakables lists.
	os.site.RLock()
	defer os.site.RUnlock()
	siteClone := os.site.Copy()
	OriginalTweakables, _ := os.site.GetTweakablesAndResiduals()
	CloneTweakables, _ := siteClone.GetTweakablesAndResiduals()

	done := make(chan struct{})

	// Copy the parameters from siteClone to the original site.
	uiSync := func() {
		// Lock clone.
		siteClone.RLock()
		defer siteClone.RUnlock()

		// Lock original site/event environment.
		event.EventEnv().Lock()
		defer event.EventEnv().UnlockRender()

		// Overwrite the value of all globalSite tweakables with siteClone tweakables.
		for i, tweakable := range CloneTweakables {
			OriginalTweakables[i].SetTweakableValue(tweakable.TweakableValue())
		}
	}

	checkStop := func() bool {
		os.RLock()
		defer os.RUnlock()

		return os.stopFlag
	}

	// Call uiSync every now and then until the optimization is done.
	go func() {
		ticker := time.NewTicker(250 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-done:
				uiSync()
				return
			case <-ticker.C:
				uiSync()
			}
		}
	}()

	// Optimize.
	go func() {
		err := Optimize(siteClone, checkStop)
		if err != nil {
			log.Printf("Optimize failed: %v", err)
		}

		os.Lock()
		defer os.Unlock()
		os.running = false

		// Stop UI updates.
		done <- struct{}{}
	}()
}

func (os *OptimizerState) Stop() {
	os.Lock()
	defer os.Unlock()

	os.stopFlag = true
}
