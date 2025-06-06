// Copyright Valkey GLIDE Project Contributors - SPDX Identifier: Apache-2.0

package integTest

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/valkey-io/valkey-glide/go/api"
	"github.com/valkey-io/valkey-glide/go/api/errors"
	"github.com/valkey-io/valkey-glide/go/api/options"

	"github.com/stretchr/testify/assert"
)

func (suite *GlideTestSuite) TestCustomCommandInfo() {
	client := suite.defaultClient()
	result, err := client.CustomCommand([]string{"INFO"})

	assert.Nil(suite.T(), err)
	assert.IsType(suite.T(), "", result)
	strResult := result.(string)
	assert.True(suite.T(), strings.Contains(strResult, "# Stats"))
}

func (suite *GlideTestSuite) TestCustomCommandPing_StringResponse() {
	client := suite.defaultClient()
	result, err := client.CustomCommand([]string{"PING"})

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "PONG", result.(string))
}

func (suite *GlideTestSuite) TestCustomCommandClientInfo() {
	clientName := "TEST_CLIENT_NAME"
	config := api.NewGlideClientConfiguration().
		WithAddress(&suite.standaloneHosts[0]).
		WithClientName(clientName)
	client := suite.client(config)

	result, err := client.CustomCommand([]string{"CLIENT", "INFO"})

	assert.Nil(suite.T(), err)
	assert.IsType(suite.T(), "", result)
	strResult := result.(string)
	assert.True(suite.T(), strings.Contains(strResult, fmt.Sprintf("name=%s", clientName)))
}

func (suite *GlideTestSuite) TestCustomCommandGet_NullResponse() {
	client := suite.defaultClient()
	key := uuid.New().String()
	result, err := client.CustomCommand([]string{"GET", key})

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), nil, result)
}

func (suite *GlideTestSuite) TestCustomCommandDel_LongResponse() {
	client := suite.defaultClient()
	key := uuid.New().String()
	suite.verifyOK(client.Set(key, "value"))
	result, err := client.CustomCommand([]string{"DEL", key})

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), int64(1), result.(int64))
}

func (suite *GlideTestSuite) TestCustomCommandHExists_BoolResponse() {
	client := suite.defaultClient()
	fields := map[string]string{"field1": "value1"}
	key := uuid.New().String()

	res1, err := client.HSet(key, fields)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), int64(1), res1)

	result, err := client.CustomCommand([]string{"HEXISTS", key, "field1"})

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), true, result.(bool))
}

func (suite *GlideTestSuite) TestCustomCommandIncrByFloat_FloatResponse() {
	client := suite.defaultClient()
	key := uuid.New().String()

	result, err := client.CustomCommand([]string{"INCRBYFLOAT", key, fmt.Sprintf("%f", 0.1)})

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), float64(0.1), result.(float64))
}

func (suite *GlideTestSuite) TestCustomCommandMGet_ArrayResponse() {
	clientName := "TEST_CLIENT_NAME"
	config := api.NewGlideClientConfiguration().
		WithAddress(&suite.standaloneHosts[0]).
		WithClientName(clientName)
	client := suite.client(config)

	key1 := uuid.New().String()
	key2 := uuid.New().String()
	key3 := uuid.New().String()
	oldValue := uuid.New().String()
	value := uuid.New().String()
	suite.verifyOK(client.Set(key1, oldValue))
	keyValueMap := map[string]string{
		key1: value,
		key2: value,
	}
	suite.verifyOK(client.MSet(keyValueMap))
	keys := []string{key1, key2, key3}
	values := []interface{}{value, value, nil}
	result, err := client.CustomCommand(append([]string{"MGET"}, keys...))

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), values, result.([]interface{}))
}

func (suite *GlideTestSuite) TestCustomCommandConfigGet_MapResponse() {
	client := suite.defaultClient()

	if suite.serverVersion < "7.0.0" {
		suite.T().Skip("This feature is added in version 7")
	}
	configMap := map[string]string{"timeout": "1000", "maxmemory": "1GB"}
	suite.verifyOK(client.ConfigSet(configMap))

	result2, err := client.CustomCommand([]string{"CONFIG", "GET", "timeout", "maxmemory"})
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), map[string]interface{}{"timeout": "1000", "maxmemory": "1073741824"}, result2)
}

func (suite *GlideTestSuite) TestCustomCommandConfigSMembers_SetResponse() {
	client := suite.defaultClient()
	key := uuid.NewString()
	members := []string{"member1", "member2", "member3"}

	res1, err := client.SAdd(key, members)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), int64(3), res1)

	result2, err := client.CustomCommand([]string{"SMEMBERS", key})
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), map[string]struct{}{"member1": {}, "member2": {}, "member3": {}}, result2)
}

func (suite *GlideTestSuite) TestCustomCommand_invalidCommand() {
	client := suite.defaultClient()
	result, err := client.CustomCommand([]string{"pewpew"})

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), err)
	assert.IsType(suite.T(), &errors.RequestError{}, err)
}

func (suite *GlideTestSuite) TestCustomCommand_invalidArgs() {
	client := suite.defaultClient()
	result, err := client.CustomCommand([]string{"ping", "pang", "pong"})

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), err)
	assert.IsType(suite.T(), &errors.RequestError{}, err)
}

func (suite *GlideTestSuite) TestCustomCommand_closedClient() {
	client := suite.defaultClient()
	client.Close()

	result, err := client.CustomCommand([]string{"ping"})

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), err)
	assert.IsType(suite.T(), &errors.ClosingError{}, err)
}

func (suite *GlideTestSuite) TestConfigSetAndGet_multipleArgs() {
	client := suite.defaultClient()

	if suite.serverVersion < "7.0.0" {
		suite.T().Skip("This feature is added in version 7")
	}
	configMap := map[string]string{"timeout": "1000", "maxmemory": "1GB"}
	resultConfigMap := map[string]string{"timeout": "1000", "maxmemory": "1073741824"}
	suite.verifyOK(client.ConfigSet(configMap))

	result2, err := client.ConfigGet([]string{"timeout", "maxmemory"})
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), resultConfigMap, result2)
}

func (suite *GlideTestSuite) TestConfigSetAndGet_noArgs() {
	client := suite.defaultClient()

	configMap := map[string]string{}

	_, err := client.ConfigSet(configMap)
	assert.NotNil(suite.T(), err)
	assert.IsType(suite.T(), &errors.RequestError{}, err)

	result2, err := client.ConfigGet([]string{})
	assert.Nil(suite.T(), result2)
	assert.NotNil(suite.T(), err)
	assert.IsType(suite.T(), &errors.RequestError{}, err)
}

func (suite *GlideTestSuite) TestConfigSetAndGet_invalidArgs() {
	client := suite.defaultClient()

	configMap := map[string]string{"time": "1000"}

	_, err := client.ConfigSet(configMap)
	assert.NotNil(suite.T(), err)
	assert.IsType(suite.T(), &errors.RequestError{}, err)

	result2, err := client.ConfigGet([]string{"time"})
	assert.Equal(suite.T(), map[string]string{}, result2)
	assert.Nil(suite.T(), err)
}

func (suite *GlideTestSuite) TestSelect_WithValidIndex() {
	client := suite.defaultClient()
	index := int64(1)
	suite.verifyOK(client.Select(index))

	key := uuid.New().String()
	value := uuid.New().String()
	suite.verifyOK(client.Set(key, value))

	res, err := client.Get(key)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), value, res.Value())
}

func (suite *GlideTestSuite) TestSelect_InvalidIndex_OutOfBounds() {
	client := suite.defaultClient()

	result, err := client.Select(-1)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), "", result)

	result, err = client.Select(1000)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), "", result)
}

func (suite *GlideTestSuite) TestSelect_SwitchBetweenDatabases() {
	client := suite.defaultClient()

	key1 := uuid.New().String()
	value1 := uuid.New().String()
	suite.verifyOK(client.Select(0))
	suite.verifyOK(client.Set(key1, value1))

	key2 := uuid.New().String()
	value2 := uuid.New().String()
	suite.verifyOK(client.Select(1))
	suite.verifyOK(client.Set(key2, value2))

	result, err := client.Get(key1)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "", result.Value())

	suite.verifyOK(client.Select(0))
	result, err = client.Get(key2)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "", result.Value())

	suite.verifyOK(client.Select(1))
	result, err = client.Get(key2)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), value2, result.Value())
}

func (suite *GlideTestSuite) TestSortReadOnlyWithOptions_ExternalWeights() {
	client := suite.defaultClient()
	if suite.serverVersion < "7.0.0" {
		suite.T().Skip("This feature is added in version 7")
	}
	key := uuid.New().String()
	client.LPush(key, []string{"item1", "item2", "item3"})

	client.Set("weight_item1", "3")
	client.Set("weight_item2", "1")
	client.Set("weight_item3", "2")

	options := options.NewSortOptions().
		SetByPattern("weight_*").
		SetOrderBy(options.ASC).
		SetIsAlpha(false)

	sortResult, err := client.SortReadOnlyWithOptions(key, *options)

	assert.Nil(suite.T(), err)
	resultList := []api.Result[string]{
		api.CreateStringResult("item2"),
		api.CreateStringResult("item3"),
		api.CreateStringResult("item1"),
	}
	assert.Equal(suite.T(), resultList, sortResult)
}

func (suite *GlideTestSuite) TestSortReadOnlyWithOptions_GetPatterns() {
	client := suite.defaultClient()
	if suite.serverVersion < "7.0.0" {
		suite.T().Skip("This feature is added in version 7")
	}
	key := uuid.New().String()
	client.LPush(key, []string{"item1", "item2", "item3"})

	client.Set("object_item1", "Object_1")
	client.Set("object_item2", "Object_2")
	client.Set("object_item3", "Object_3")

	options := options.NewSortOptions().
		SetByPattern("weight_*").
		SetOrderBy(options.ASC).
		SetIsAlpha(false).
		AddGetPattern("object_*")

	sortResult, err := client.SortReadOnlyWithOptions(key, *options)

	assert.Nil(suite.T(), err)

	resultList := []api.Result[string]{
		api.CreateStringResult("Object_2"),
		api.CreateStringResult("Object_3"),
		api.CreateStringResult("Object_1"),
	}

	assert.Equal(suite.T(), resultList, sortResult)
}

func (suite *GlideTestSuite) TestSortReadOnlyWithOptions_SuccessfulSortByWeightAndGet() {
	client := suite.defaultClient()
	if suite.serverVersion < "7.0.0" {
		suite.T().Skip("This feature is added in version 7")
	}
	key := uuid.New().String()
	client.LPush(key, []string{"item1", "item2", "item3"})

	client.Set("weight_item1", "10")
	client.Set("weight_item2", "5")
	client.Set("weight_item3", "15")

	client.Set("object_item1", "Object 1")
	client.Set("object_item2", "Object 2")
	client.Set("object_item3", "Object 3")

	options := options.NewSortOptions().
		SetOrderBy(options.ASC).
		SetIsAlpha(false).
		SetByPattern("weight_*").
		AddGetPattern("object_*").
		AddGetPattern("#")

	sortResult, err := client.SortReadOnlyWithOptions(key, *options)

	assert.Nil(suite.T(), err)

	resultList := []api.Result[string]{
		api.CreateStringResult("Object 2"),
		api.CreateStringResult("item2"),
		api.CreateStringResult("Object 1"),
		api.CreateStringResult("item1"),
		api.CreateStringResult("Object 3"),
		api.CreateStringResult("item3"),
	}

	assert.Equal(suite.T(), resultList, sortResult)

	objectItem2, err := client.Get("object_item2")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "Object 2", objectItem2.Value())

	objectItem1, err := client.Get("object_item1")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "Object 1", objectItem1.Value())

	objectItem3, err := client.Get("object_item3")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "Object 3", objectItem3.Value())

	assert.Equal(suite.T(), "item2", sortResult[1].Value())
	assert.Equal(suite.T(), "item1", sortResult[3].Value())
	assert.Equal(suite.T(), "item3", sortResult[5].Value())
}

func (suite *GlideTestSuite) TestInfoStandalone() {
	DEFAULT_INFO_SECTIONS := []string{
		"Server",
		"Clients",
		"Memory",
		"Persistence",
		"Stats",
		"Replication",
		"CPU",
		"Modules",
		"Errorstats",
		"Cluster",
		"Keyspace",
	}

	client := suite.defaultClient()
	t := suite.T()

	// info without options
	info, err := client.Info()
	assert.NoError(t, err)
	for _, section := range DEFAULT_INFO_SECTIONS {
		assert.Contains(t, info, "# "+section, "Section "+section+" is missing")
	}

	// info with option or with multiple options
	sections := []options.Section{options.Cpu}
	if suite.serverVersion >= "7.0.0" {
		sections = append(sections, options.Memory)
	}
	info, err = client.InfoWithOptions(options.InfoOptions{Sections: sections})
	assert.NoError(t, err)
	for _, section := range sections {
		assert.Contains(t, strings.ToLower(info), strings.ToLower("# "+string(section)), "Section "+section+" is missing")
	}
}

func (suite *GlideTestSuite) TestDBSize() {
	client := suite.defaultClient()
	result, err := client.DBSize()
	assert.Nil(suite.T(), err)
	assert.Greater(suite.T(), result, int64(0))
}

func (suite *GlideTestSuite) TestPing_NoArgument() {
	client := suite.defaultClient()

	result, err := client.Ping()
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "PONG", result)
}

func (suite *GlideTestSuite) TestEcho() {
	client := suite.defaultClient()
	// Test 1: Check if Echo command return the message
	value := "Hello world"
	t := suite.T()
	resultEcho, err := client.Echo(value)
	assert.Nil(t, err)
	assert.Equal(t, value, resultEcho.Value())
}

func (suite *GlideTestSuite) TestPing_ClosedClient() {
	client := suite.defaultClient()
	client.Close()

	result, err := client.Ping()

	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), "", result)
	assert.IsType(suite.T(), &errors.ClosingError{}, err)
}

func (suite *GlideTestSuite) TestPingWithOptions_WithMessage() {
	client := suite.defaultClient()
	options := options.PingOptions{
		Message: "hello",
	}

	result, err := client.PingWithOptions(options)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "hello", result)
}

func (suite *GlideTestSuite) TestPingWithOptions_ClosedClient() {
	client := suite.defaultClient()
	client.Close()

	options := options.PingOptions{
		Message: "hello",
	}

	result, err := client.PingWithOptions(options)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), "", result)
	assert.IsType(suite.T(), &errors.ClosingError{}, err)
}

func (suite *GlideTestSuite) TestTime_Success() {
	client := suite.defaultClient()
	results, err := client.Time()

	assert.Nil(suite.T(), err)
	assert.Len(suite.T(), results, 2)

	now := time.Now().Unix() - 1

	timestamp, err := strconv.ParseInt(results[0], 10, 64)
	assert.Nil(suite.T(), err)
	assert.Greater(suite.T(), timestamp, now)

	microseconds, err := strconv.ParseInt(results[1], 10, 64)
	assert.Nil(suite.T(), err)
	assert.Less(suite.T(), microseconds, int64(1000000))
}

func (suite *GlideTestSuite) TestTime_Error() {
	client := suite.defaultClient()

	// Disconnect the client or simulate an error condition
	client.Close()

	results, err := client.Time()

	assert.NotNil(suite.T(), err)
	assert.Nil(suite.T(), results)
	assert.IsType(suite.T(), &errors.ClosingError{}, err)
}

func (suite *GlideTestSuite) TestFlushAll() {
	client := suite.defaultClient()
	key1 := uuid.New().String()
	key2 := uuid.New().String()

	_, err := client.Set(key1, "value1")
	assert.Nil(suite.T(), err)
	_, err = client.Set(key2, "value2")
	assert.Nil(suite.T(), err)

	result, err := client.Get(key1)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "value1", result.Value())

	response, err := client.FlushAll()
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "OK", response)

	result, err = client.Get(key1)
	assert.Nil(suite.T(), err)
	assert.Empty(suite.T(), result.Value())
}

func (suite *GlideTestSuite) TestFlushAll_Sync() {
	client := suite.defaultClient()
	key1 := uuid.New().String()
	key2 := uuid.New().String()

	_, err := client.Set(key1, "value1")
	assert.Nil(suite.T(), err)
	_, err = client.Set(key2, "value2")
	assert.Nil(suite.T(), err)

	result, err := client.Get(key1)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "value1", result.Value())

	response, err := client.FlushAllWithOptions(options.SYNC)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "OK", response)

	result, err = client.Get(key1)
	assert.Nil(suite.T(), err)
	assert.Empty(suite.T(), result.Value())
}

func (suite *GlideTestSuite) TestFlushAll_Async() {
	client := suite.defaultClient()
	key1 := uuid.New().String()
	key2 := uuid.New().String()

	_, err := client.Set(key1, "value1")
	assert.Nil(suite.T(), err)
	_, err = client.Set(key2, "value2")
	assert.Nil(suite.T(), err)

	response, err := client.FlushAllWithOptions(options.ASYNC)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "OK", response)

	result, err := client.Get(key1)
	assert.Nil(suite.T(), err)
	assert.Empty(suite.T(), result.Value())
}

func (suite *GlideTestSuite) TestFlushAll_ClosedClient() {
	client := suite.defaultClient()
	client.Close()

	response, err := client.FlushAllWithOptions(options.SYNC)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), "", response)
	assert.IsType(suite.T(), &errors.ClosingError{}, err)
}

func (suite *GlideTestSuite) TestFlushAll_MultipleFlush() {
	client := suite.defaultClient()
	key1 := uuid.New().String()

	response, err := client.FlushAllWithOptions(options.SYNC)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "OK", response)

	_, err = client.Set(key1, "value1")
	assert.Nil(suite.T(), err)

	response, err = client.FlushAllWithOptions(options.ASYNC)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "OK", response)

	result, err := client.Get(key1)
	assert.Nil(suite.T(), err)
	assert.Empty(suite.T(), result.Value())
}

func (suite *GlideTestSuite) TestFlushDB() {
	client := suite.defaultClient()
	key1 := uuid.New().String()
	key2 := uuid.New().String()

	_, err := client.Set(key1, "value1")
	assert.Nil(suite.T(), err)
	_, err = client.Set(key2, "value2")
	assert.Nil(suite.T(), err)

	result, err := client.Get(key1)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "value1", result.Value())

	response, err := client.FlushDB()
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "OK", response)

	result, err = client.Get(key1)
	assert.Nil(suite.T(), err)
	assert.Empty(suite.T(), result.Value())
}

func (suite *GlideTestSuite) TestFlushDBWithOptions_SYNC() {
	client := suite.defaultClient()

	key := uuid.New().String()
	_, err := client.Set(key, "value1")
	assert.NoError(suite.T(), err)

	result, err := client.FlushDBWithOptions(options.SYNC)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), result)

	val, err := client.Get(key)
	assert.NoError(suite.T(), err)
	assert.Empty(suite.T(), val.Value())
}

func (suite *GlideTestSuite) TestFlushDBWithOptions_ASYNC() {
	client := suite.defaultClient()

	key := uuid.New().String()
	_, err := client.Set(key, "value1")
	assert.NoError(suite.T(), err)

	result, err := client.FlushDBWithOptions(options.ASYNC)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), result)

	val, err := client.Get(key)
	assert.NoError(suite.T(), err)
	assert.Empty(suite.T(), val.Value())
}

func (suite *GlideTestSuite) TestFlushDBWithOptions_MultipleKeys() {
	client := suite.defaultClient()

	key1 := uuid.New().String()
	key2 := uuid.New().String()

	_, err := client.Set(key1, "value1")
	assert.NoError(suite.T(), err)
	_, err = client.Set(key2, "value2")
	assert.NoError(suite.T(), err)

	result, err := client.FlushDBWithOptions(options.SYNC)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), result)

	val1, err := client.Get(key1)
	assert.NoError(suite.T(), err)
	assert.Empty(suite.T(), val1.Value())

	val2, err := client.Get(key2)
	assert.NoError(suite.T(), err)
	assert.Empty(suite.T(), val2.Value())
}

func (suite *GlideTestSuite) TestFlushDBWithOptions_ClosedClient() {
	client := suite.defaultClient()

	client.Close()

	result, err := client.FlushDBWithOptions(options.SYNC)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), "", result)
	assert.IsType(suite.T(), &errors.ClosingError{}, err)
}

func (suite *GlideTestSuite) TestUpdateConnectionPasswordAuthNonValidPass() {
	// Create test client
	testClient := suite.defaultClient()
	defer testClient.Close()

	// Test empty password
	_, err := testClient.UpdateConnectionPassword("", true)
	assert.NotNil(suite.T(), err)
	assert.IsType(suite.T(), &errors.RequestError{}, err)

	// Test with no password parameter
	_, err = testClient.UpdateConnectionPassword("", true)
	assert.NotNil(suite.T(), err)
	assert.IsType(suite.T(), &errors.RequestError{}, err)
}

func (suite *GlideTestSuite) TestUpdateConnectionPassword_NoServerAuth() {
	// Create test client
	testClient := suite.defaultClient()
	defer testClient.Close()

	// Validate that we can use the client
	_, err := testClient.Info()
	assert.Nil(suite.T(), err)

	// Test immediate re-authentication fails when no server password is set
	pwd := uuid.NewString()
	_, err = testClient.UpdateConnectionPassword(pwd, true)
	assert.NotNil(suite.T(), err)
	assert.IsType(suite.T(), &errors.RequestError{}, err)
}

func (suite *GlideTestSuite) TestUpdateConnectionPassword_LongPassword() {
	// Create test client
	testClient := suite.defaultClient()
	defer testClient.Close()

	// Generate long random password (1000 chars)
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	pwd := make([]byte, 1000)
	for i := range pwd {
		pwd[i] = letters[rand.Intn(len(letters))]
	}

	// Validate that we can use the client
	_, err := testClient.Info()
	assert.NoError(suite.T(), err)

	// Test replacing connection password with a long password string
	_, err = testClient.UpdateConnectionPassword(string(pwd), false)
	assert.NoError(suite.T(), err)
}

func (suite *GlideTestSuite) TestUpdateConnectionPassword_ImmediateAuthWrongPassword() {
	// Create admin client
	adminClient := suite.defaultClient()
	defer adminClient.Close()

	// Create test client
	testClient := suite.defaultClient()
	defer testClient.Close()

	pwd := uuid.NewString()
	notThePwd := uuid.NewString()

	// Validate that we can use the client
	_, err := testClient.Info()
	assert.Nil(suite.T(), err)

	// Set the password to something else
	_, err = adminClient.ConfigSet(map[string]string{"requirepass": notThePwd})
	assert.Nil(suite.T(), err)

	// Test that re-authentication fails when using wrong password
	_, err = testClient.UpdateConnectionPassword(pwd, true)
	assert.NotNil(suite.T(), err)
	assert.IsType(suite.T(), &errors.RequestError{}, err)

	// But using correct password returns OK
	_, err = testClient.UpdateConnectionPassword(notThePwd, true)
	assert.NoError(suite.T(), err)

	// Cleanup: Reset password
	_, err = adminClient.ConfigSet(map[string]string{"requirepass": ""})
	assert.NoError(suite.T(), err)
}

func (suite *GlideTestSuite) TestLolwutWithOptions_WithVersion() {
	client := suite.defaultClient()
	options := options.NewLolwutOptions(8)
	res, err := client.LolwutWithOptions(*options)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), res, "Redis ver.")
}

func (suite *GlideTestSuite) TestLolwutWithOptions_WithVersionAndArgs() {
	client := suite.defaultClient()
	opts := options.NewLolwutOptions(8).SetArgs([]int{10, 20})
	res, err := client.LolwutWithOptions(*opts)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), res, "Redis ver.")
}

func (suite *GlideTestSuite) TestLolwutWithOptions_EmptyArgs() {
	client := suite.defaultClient()
	opts := options.NewLolwutOptions(6).SetArgs([]int{})
	res, err := client.LolwutWithOptions(*opts)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), res, "Redis ver.")
}

func (suite *GlideTestSuite) TestClientId() {
	client := suite.defaultClient()
	result, err := client.ClientId()
	assert.Nil(suite.T(), err)
	assert.Greater(suite.T(), result, int64(0))
}

func (suite *GlideTestSuite) TestLastSave() {
	client := suite.defaultClient()
	t := suite.T()
	result, err := client.LastSave()
	assert.Nil(t, err)
	assert.Greater(t, result, int64(0))
}

func (suite *GlideTestSuite) TestConfigResetStat() {
	client := suite.defaultClient()
	suite.verifyOK(client.ConfigResetStat())
}

func (suite *GlideTestSuite) TestClientGetSetName() {
	client := suite.defaultClient()
	t := suite.T()

	suite.verifyOK(client.ClientSetName("ConnectionName"))
	result, err := client.ClientGetName()
	assert.Nil(t, err)
	assert.Equal(t, result, "ConnectionName")
}

func (suite *GlideTestSuite) TestMove() {
	client := suite.defaultClient()
	t := suite.T()
	key := uuid.New().String()
	suite.verifyOK(client.Set(key, "hello"))
	result, err := client.Move(key, 2)
	assert.Nil(t, err)
	assert.True(suite.T(), result)
}

func (suite *GlideTestSuite) TestScan() {
	client := suite.defaultClient()
	t := suite.T()
	key := uuid.New().String()
	suite.verifyOK(client.Set(key, "Hello"))
	resCursor, resCollection, err := client.Scan(0)
	assert.Nil(t, err)
	assert.GreaterOrEqual(t, len(resCursor), 1)
	assert.GreaterOrEqual(t, len(resCollection), 1)
}

func (suite *GlideTestSuite) TestScanWithOption() {
	client := suite.defaultClient()
	t := suite.T()

	// Test TestScanWithOption SetCount
	key := uuid.New().String()
	suite.verifyOK(client.Set(key, "Hello"))
	opts := options.NewScanOptions().SetCount(10)
	resCursor, resCollection, err := client.ScanWithOptions(0, *opts)
	assert.Nil(t, err)
	assert.GreaterOrEqual(t, len(resCursor), 1)
	assert.GreaterOrEqual(t, len(resCollection), 1)

	// Test TestScanWithOption SetType
	opts = options.NewScanOptions().SetType(options.ObjectTypeString)
	resCursor, resCollection, err = client.ScanWithOptions(0, *opts)
	assert.Nil(t, err)
	assert.GreaterOrEqual(t, len(resCursor), 1)
	assert.GreaterOrEqual(t, len(resCollection), 1)
}

func (suite *GlideTestSuite) TestConfigRewrite() {
	client := suite.defaultClient()
	t := suite.T()
	opts := options.InfoOptions{Sections: []options.Section{options.Server}}
	response, err := client.InfoWithOptions(opts)
	assert.NoError(t, err)
	lines := strings.Split(response, "\n")
	var configFile string
	for _, line := range lines {
		if strings.HasPrefix(line, "config_file:") {
			configFile = strings.TrimSpace(strings.TrimPrefix(line, "config_file:"))
			break
		}
	}
	if len(configFile) > 0 {
		suite.verifyOK(client.ConfigRewrite())
	}
}

func (suite *GlideTestSuite) TestRandomKey() {
	client := suite.defaultClient()
	// Test 1: Check if the command return random key
	t := suite.T()
	result, err := client.RandomKey()
	assert.Nil(t, err)
	assert.NotNil(t, result)
}

func (suite *GlideTestSuite) TestFunctionCommandsStandalone() {
	if suite.serverVersion < "7.0.0" {
		suite.T().Skip("This feature is added in version 7")
	}

	client := suite.defaultClient()

	// Flush all functions with SYNC option
	result, err := client.FunctionFlushSync()
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "OK", result)

	// Generate and load function
	libName := "mylib1c"
	funcName := "myfunc1c"
	functions := map[string]string{
		funcName: "return args[1]",
	}
	code := GenerateLuaLibCode(libName, functions, true)
	result, err = client.FunctionLoad(code, false)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), libName, result)

	// Test FCALL
	functionResult, err := client.FCallWithKeysAndArgs(funcName, []string{}, []string{"one", "two"})
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "one", functionResult)

	// Test FCALL_RO
	functionResult, err = client.FCallReadOnlyWithKeysAndArgs(funcName, []string{}, []string{"one", "two"})
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "one", functionResult)
}
