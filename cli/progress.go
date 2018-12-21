// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2018 Roberto Mier Escandon <rmescandon@gmail.com>
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/greenbrew/rest/cli/progress"
)

// ShowProgressSpin prints out a rotating spin in console while the op function is being executed
func ShowProgressSpin(ctx context.Context, summary string, op func(ctx context.Context) error) error {
	pb := progress.MakeProgressBar()
	defer func() {
		pb.Finished()
	}()

	ch := make(chan error, 1)
	go func() {
		err := op(ctx)
		ch <- err
	}()

	for {
		select {
		case err, ok := <-ch:
			if ok {
				return err
			}
			return fmt.Errorf("Operation failed")
		default:
			break
		}

		pb.Spin(summary)
		time.Sleep(time.Millisecond * 200)
	}
}

// ShowProgressBar prints out a progress bar in console while op function is being executed.
// A 'max' value and 'sent' channel parameters are needed to visually progress the bar
func ShowProgressBar(ctx context.Context, summary string, max float64, sent chan float64, op func(ctx context.Context) error) error {
	pb := prepareProgressBar(summary, max)
	defer func() {
		pb.Finished()
	}()

	return performOperationAndShowProgress(ctx, pb, sent, op)
}

func prepareProgressBar(summary string, max float64) progress.Meter {
	pb := progress.MakeProgressBar()
	pb.Start(summary, max)
	return pb
}

func performOperationAndShowProgress(ctx context.Context, pb progress.Meter, sent chan float64, op func(ctx context.Context) error) error {
	ch := make(chan error, 1)
	go func() {
		err := op(ctx)
		ch <- err
	}()

	var total float64

	for {
		select {
		case err, ok := <-ch:
			if ok {
				return err
			}
			return fmt.Errorf("Operation failed")
		case n := <-sent:
			total += n
			pb.Set(total)
		}
	}
}
