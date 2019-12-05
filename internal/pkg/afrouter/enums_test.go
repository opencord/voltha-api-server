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

package afrouter

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

//function to test Marshal and UnMarshal JSON of backend Type
func Test_backendType(t *testing.T) {
	t_backendType1 := backendType(1)
	backend_Str1, err := t_backendType1.MarshalJSON()
	assert.Equal(t, string(backend_Str1), "\"active_active\"")
	assert.Equal(t, err, nil)

	t_backendType2 := backendType(2)
	backend_Str2, err := t_backendType2.MarshalJSON()
	assert.Equal(t, string(backend_Str2), "\"server\"")
	assert.Equal(t, err, nil)

	t_backendStr1 := backendType(1)
	backend_Type1 := t_backendStr1.UnmarshalJSON([]byte("\"active_active\""))
	assert.Equal(t, backend_Type1, nil)

	t_backendStr2 := backendType(2)
	backend_Type2 := t_backendStr2.UnmarshalJSON([]byte("\"server\""))
	assert.Equal(t, backend_Type2, nil)

}

//function to test Marshal and UnMarshal JSON of association Location
func Test_AssociationLocation(t *testing.T) {
	t_assocLoc1 := associationLocation(1)
	assoc_Str1, err := t_assocLoc1.MarshalJSON()
	assert.Equal(t, string(assoc_Str1), "\"header\"")
	assert.Equal(t, err, nil)

	t_assocLoc2 := associationLocation(2)
	assoc_Str2, err := t_assocLoc2.MarshalJSON()
	assert.Equal(t, string(assoc_Str2), "\"protobuf\"")
	assert.Equal(t, err, nil)

	t_assocStr1 := associationLocation(1)
	assoc_Loc1 := t_assocStr1.UnmarshalJSON([]byte("\"header\""))
	assert.Equal(t, assoc_Loc1, nil)

	t_assocStr2 := associationLocation(2)
	assoc_Loc2 := t_assocStr2.UnmarshalJSON([]byte("\"protobuf\""))
	assert.Equal(t, assoc_Loc2, nil)
}

//function to test Marshal and UnMarshal JSON of assoc Loc strategy
func Test_AssociationStrategy(t *testing.T) {
	t_assocStrat1 := associationStrategy(1)
	assocstrat_Str1, err := t_assocStrat1.MarshalJSON()
	assert.Equal(t, string(assocstrat_Str1), "\"serial_number\"")
	assert.Equal(t, err, nil)

	t_assocstratStr1 := associationStrategy(1)
	assocstrat1 := t_assocstratStr1.UnmarshalJSON([]byte("\"serial_number\""))
	assert.Equal(t, assocstrat1, nil)
}

//function to test Marshal and UnMarshal JSON of route type
func Test_RouteType(t *testing.T) {
	t_routetyp1 := routeType(1)
	routetyp_Str1, err := t_routetyp1.MarshalJSON()
	assert.Equal(t, string(routetyp_Str1), "\"rpc_affinity_message\"")
	assert.Equal(t, err, nil)

	t_routetyp2 := routeType(2)
	routetyp_Str2, err := t_routetyp2.MarshalJSON()
	assert.Equal(t, string(routetyp_Str2), "\"rpc_affinity_header\"")
	assert.Equal(t, err, nil)

	t_routetyp3 := routeType(3)
	routetyp_Str3, err := t_routetyp3.MarshalJSON()
	assert.Equal(t, string(routetyp_Str3), "\"binding\"")
	assert.Equal(t, err, nil)

	t_routetyp4 := routeType(4)
	routetyp_Str4, err := t_routetyp4.MarshalJSON()
	assert.Equal(t, string(routetyp_Str4), "\"round_robin\"")
	assert.Equal(t, err, nil)

	t_routetyp5 := routeType(5)
	routetyp_Str5, err := t_routetyp5.MarshalJSON()
	assert.Equal(t, string(routetyp_Str5), "\"source\"")
	assert.Equal(t, err, nil)

	t_routeStr1 := routeType(1)
	routeType1 := t_routeStr1.UnmarshalJSON([]byte("\"rpc_affinity_message\""))
	assert.Equal(t, routeType1, nil)

	t_routeStr2 := routeType(2)
	routeType2 := t_routeStr2.UnmarshalJSON([]byte("\"rpc_affinity_header\""))
	assert.Equal(t, routeType2, nil)

	t_routeStr3 := routeType(3)
	routeType3 := t_routeStr3.UnmarshalJSON([]byte("\"binding\""))
	assert.Equal(t, routeType3, nil)

	t_routeStr4 := routeType(4)
	routeType4 := t_routeStr4.UnmarshalJSON([]byte("\"round_robin\""))
	assert.Equal(t, routeType4, nil)

	t_routeStr5 := routeType(5)
	routeType5 := t_routeStr5.UnmarshalJSON([]byte("\"source\""))
	assert.Equal(t, routeType5, nil)
}

//function to test Marshal and UnMarshal JSON of Assoc type
func Test_AssociationType(t *testing.T) {
	t_assocType1 := associationType(1)
	assocStr1, err := t_assocType1.MarshalJSON()
	assert.Equal(t, string(assocStr1), "\"round_robin\"")
	assert.Equal(t, err, nil)

	t_assocStr1 := associationType(1)
	assocType1 := t_assocStr1.UnmarshalJSON([]byte("\"round_robin\""))
	assert.Equal(t, assocType1, nil)
}
