/*
 * Copyright (C) 2021 The "MysteriumNetwork/node" Authors.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package behavior

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// ErrEmptyAddressList indicates there are no servers to get response from
var ErrEmptyAddressList = errors.New("empty STUN server list specified")

type discoverResult struct {
	res string
	err error
}

func RacingDiscoverNATBehavior(ctx context.Context, addresses []string, timeout time.Duration) (string, error) {
	count := len(addresses)

	ctx1, cl := context.WithCancel(ctx)
	defer cl()

	results := make(chan discoverResult)

	for _, address := range addresses {
		go func(address string) {
			res, err := DiscoverNATBehavior(ctx1, address, timeout)
			resPair := discoverResult{res, err}
			select {
			case results <- resPair:
			case <-ctx1.Done():
			}
		}(address)
	}

	lastError := ErrEmptyAddressList
	for i := 0; i < count; i++ {
		select {
		case res := <-results:
			if res.err != nil {
				lastError = res.err
			} else {
				return res.res, res.err
			}
		case <-ctx1.Done():
			return "", ctx1.Err()
		}
	}

	return "", fmt.Errorf("concurrent NAT probing failed. last error: %w", lastError)
}
