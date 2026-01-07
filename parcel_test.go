package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	id, err:= store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)
	parcel.Number = id

	// get
	getParcel, err:= store.Get(parcel.Number)
	require.NoError(t, err)
	require.Equal(t, parcel.Client, getParcel.Client)
	require.Equal(t, parcel.Status, getParcel.Status)
	require.Equal(t, parcel.Address, getParcel.Address)
	require.Equal(t, parcel.CreatedAt, getParcel.CreatedAt)

	// delete
	err = store.Delete(parcel.Number)
	require.NoError(t, err)
	_, err = store.Get(parcel.Number)
	require.Error(t, err)
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	id, err:= store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)
	parcel.Number = id

	// set address
	newAddress := "new test address"
	err = store.SetAddress(parcel.Number, newAddress)
	require.NoError(t, err)

	// check
	updatedParcel, err := store.Get(parcel.Number)
	require.NoError(t, err)
	require.Equal(t, newAddress, updatedParcel.Address)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	id, err:= store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)
	parcel.Number = id

	// set status
	newStatus := ParcelStatusSent
	err = store.SetStatus(parcel.Number, newStatus)
	require.NoError(t, err)

	// check
	updParcel, err := store.Get(parcel.Number)
	require.NoError(t, err)
	require.Equal(t, newStatus, updParcel.Status)
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// add
	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i])
		require.NoError(t, err)
		require.NotZero(t, id)

		parcels[i].Number = id

		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client)
	require.NoError(t, err)
	require.Len(t, storedParcels, len(parcels))

	// check
	for _, parcel := range storedParcels {
		expec, ok := parcelMap[parcel.Number]
		require.True(t, ok, "неизвестная посылка с номером %d", parcel.Number)

		require.Equal(t, expec.Client, parcel.Client)
		require.Equal(t, expec.Status, parcel.Status)
		require.Equal(t, expec.Address, parcel.Address)
		require.Equal(t, expec.CreatedAt, parcel.CreatedAt)
	}
}
