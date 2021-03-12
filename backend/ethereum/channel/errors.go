// Copyright 2019 - See NOTICE file for copyright holders.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package channel

import (
	"fmt"

	"github.com/pkg/errors"
)

type (
	// TxTimedoutError indicates that we have timed out waiting for a
	// transaction to be mined.
	// It can happen that this transaction can be eventually mined. So the user
	// should of the framework should be watching for the transaction to be
	// mined and decide on what to do.
	TxTimedoutError struct {
		TxID   string
		TxName string
	}
)

// Error implements the error interface.
func (e TxTimedoutError) Error() string {
	return fmt.Sprintf("timed out waiting for tx to be mined. txID: %s, txName: %s", e.TxID, e.TxName)
}

func newTxTimedoutError(txID, txName, msg string) error {
	return errors.Wrap(TxTimedoutError{
		TxID:   txID,
		TxName: txName,
	}, msg)
}
