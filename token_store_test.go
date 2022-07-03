package mysql_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/go-oauth2/oauth2/v4/models"

	. "github.com/httpoz/go-oauth2-mysql"
)

func generateTokenTableName() string {
	return fmt.Sprintf("token_%d", time.Now().UnixNano())
}

func generateClientTableName() string {
	return fmt.Sprintf("client_%d", time.Now().UnixNano())
}

func TestMYSQLConn(t *testing.T) {

	db, err := sqlx.Connect("mysql", os.Getenv("MYSQL_URI"))
	require.NoError(t, err)

	defer func() {
		assert.NoError(t, db.Close())
	}()

	tokenStore, err := NewTokenStore(
		db,
		WithTokenStoreTableName(generateTokenTableName()),
		WithTokenStoreGCInterval(time.Second),
	)
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, tokenStore.Close())
	}()

	clientStore, err := NewClientStore(
		db,
		WithClientStoreTableName(generateClientTableName()),
	)
	require.NoError(t, err)

	runTokenStoreTest(t, tokenStore)
	runClientStoreTest(t, clientStore)
}

func TestSQL(t *testing.T) {
	db, err := sqlx.Connect("mysql", os.Getenv("MYSQL_URI"))
	require.NoError(t, err)

	defer func() {
		assert.NoError(t, db.Close())
	}()

	tokenStore, err := NewTokenStore(
		db,
		WithTokenStoreTableName(generateTokenTableName()),
		WithTokenStoreGCInterval(time.Second),
	)
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, tokenStore.Close())
	}()

	clientStore, err := NewClientStore(
		db,
		WithClientStoreTableName(generateClientTableName()),
	)
	require.NoError(t, err)

	runTokenStoreTest(t, tokenStore)
	runClientStoreTest(t, clientStore)
}

func runTokenStoreTest(t *testing.T, store *TokenStore) {
	runTokenStoreCodeTest(t, store)
	runTokenStoreAccessTest(t, store)
	runTokenStoreRefreshTest(t, store)

	// sleep for a while just to wait for GC run for sure to ensure there were no errors there
	time.Sleep(3 * time.Second)
}

func runTokenStoreCodeTest(t *testing.T, store *TokenStore) {
	ctx := context.Background()
	code := fmt.Sprintf("code %s", time.Now().String())

	tokenCode := models.NewToken()
	tokenCode.SetCode(code)
	tokenCode.SetCodeCreateAt(time.Now())
	tokenCode.SetCodeExpiresIn(time.Minute)
	require.NoError(t, store.Create(ctx, tokenCode))

	token, err := store.GetByCode(ctx, code)
	require.NoError(t, err)
	assert.Equal(t, code, token.GetCode())

	require.NoError(t, store.RemoveByCode(ctx, code))

	_, err = store.GetByCode(ctx, code)
	assert.Nil(t, err)
}

func runTokenStoreAccessTest(t *testing.T, store *TokenStore) {
	ctx := context.Background()
	code := fmt.Sprintf("access %s", time.Now().String())

	tokenCode := models.NewToken()
	tokenCode.SetAccess(code)
	tokenCode.SetAccessCreateAt(time.Now())
	tokenCode.SetAccessExpiresIn(time.Minute)
	require.NoError(t, store.Create(ctx, tokenCode))

	token, err := store.GetByAccess(ctx, code)
	require.NoError(t, err)
	assert.Equal(t, code, token.GetAccess())

	require.NoError(t, store.RemoveByAccess(ctx, code))

	_, err = store.GetByAccess(ctx, code)
	assert.Nil(t, err)
}

func runTokenStoreRefreshTest(t *testing.T, store *TokenStore) {
	ctx := context.Background()
	code := fmt.Sprintf("refresh %s", time.Now().String())

	tokenCode := models.NewToken()
	tokenCode.SetRefresh(code)
	tokenCode.SetRefreshCreateAt(time.Now())
	tokenCode.SetRefreshExpiresIn(time.Minute)
	require.NoError(t, store.Create(ctx, tokenCode))

	token, err := store.GetByRefresh(ctx, code)
	require.NoError(t, err)
	assert.Equal(t, code, token.GetRefresh())

	require.NoError(t, store.RemoveByRefresh(ctx, code))

	_, err = store.GetByRefresh(ctx, code)
	assert.Nil(t, err)
}

func runClientStoreTest(t *testing.T, store *ClientStore) {
	ctx := context.Background()
	originalClient := &models.Client{
		ID:     fmt.Sprintf("id %s", time.Now().String()),
		Secret: fmt.Sprintf("secret %s", time.Now().String()),
		Domain: fmt.Sprintf("domain %s", time.Now().String()),
		UserID: fmt.Sprintf("user id %s", time.Now().String()),
	}

	require.NoError(t, store.Create(originalClient))

	client, err := store.GetByID(ctx, originalClient.GetID())
	require.NoError(t, err)
	assert.Equal(t, originalClient.GetID(), client.GetID())
	assert.Equal(t, originalClient.GetSecret(), client.GetSecret())
	assert.Equal(t, originalClient.GetDomain(), client.GetDomain())
	assert.Equal(t, originalClient.GetUserID(), client.GetUserID())
}
