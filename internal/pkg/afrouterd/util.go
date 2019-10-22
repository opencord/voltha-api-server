/*
 * Copyright 2018-present Open Networking Foundation

 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at

 * http://www.apache.org/licenses/LICENSE-2.0

 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package afrouterd

import (
	"fmt"
	"os"
	"strconv"
)

func GetIntEnv(key string, min, max, defaultValue int) int {
	if val, have := os.LookupEnv(key); have {
		num, err := strconv.Atoi(val)
		if err != nil || !(min <= num && num <= max) {
			panic(fmt.Errorf("%s must be a number in the range [%d, %d]; default: %d", key, min, max, defaultValue))
		}
		return num
	}
	return defaultValue
}

func GetStrEnv(key, defaultValue string) string {
	if val, have := os.LookupEnv(key); have {
		return val
	}
	return defaultValue
}
