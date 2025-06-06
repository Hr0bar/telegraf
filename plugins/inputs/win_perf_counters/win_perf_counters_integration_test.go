//go:build windows

package win_perf_counters

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/testutil"
)

func TestWinPerformanceQueryImplIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	query := &performanceQueryImpl{maxBufferSize: uint32(defaultMaxBufferSize)}

	err := query.close()
	require.Error(t, err, "uninitialized query must return errors")

	_, err = query.addCounterToQuery("")
	require.Error(t, err, "uninitialized query must return errors")
	require.ErrorContains(t, err, "uninitialized")

	_, err = query.addEnglishCounterToQuery("")
	require.Error(t, err, "uninitialized query must return errors")
	require.ErrorContains(t, err, "uninitialized")

	err = query.collectData()
	require.Error(t, err, "uninitialized query must return errors")
	require.ErrorContains(t, err, "uninitialized")

	require.NoError(t, query.open())

	counterPath := "\\Processor Information(_Total)\\% Processor Time"

	hCounter, err := query.addCounterToQuery(counterPath)
	require.NoError(t, err)
	require.NotEqual(t, 0, hCounter)

	require.NoError(t, query.close())

	require.NoError(t, query.open())

	hCounter, err = query.addEnglishCounterToQuery(counterPath)
	require.NoError(t, err)
	require.NotEqual(t, 0, hCounter)

	cp, err := query.getCounterPath(hCounter)
	require.NoError(t, err)
	require.True(t, strings.HasSuffix(cp, counterPath))

	require.NoError(t, query.collectData())
	time.Sleep(time.Second)

	require.NoError(t, query.collectData())

	fcounter, err := query.getFormattedCounterValueDouble(hCounter)
	require.NoError(t, err)
	require.Greater(t, fcounter, float64(0))

	rcounter, err := query.getRawCounterValue(hCounter)
	require.NoError(t, err)
	require.Greater(t, rcounter, int64(10000000))

	now := time.Now()
	mtime, err := query.collectDataWithTime()
	require.NoError(t, err)
	require.Less(t, mtime.Sub(now), time.Second)

	counterPath = "\\Process(*)\\% Processor Time"
	paths, err := query.expandWildCardPath(counterPath)
	require.NoError(t, err)
	require.NotNil(t, paths)
	require.Greater(t, len(paths), 1)

	counterPath = "\\Process(_Total)\\*"
	paths, err = query.expandWildCardPath(counterPath)
	require.NoError(t, err)
	require.NotNil(t, paths)
	require.Greater(t, len(paths), 1)

	require.NoError(t, query.open())

	counterPath = "\\Process(*)\\% Processor Time"
	hCounter, err = query.addEnglishCounterToQuery(counterPath)
	require.NoError(t, err)
	require.NotEqual(t, 0, hCounter)

	require.NoError(t, query.collectData())
	time.Sleep(time.Second)

	require.NoError(t, query.collectData())

	farr, err := query.getFormattedCounterArrayDouble(hCounter)
	var phdErr *pdhError
	if errors.As(err, &phdErr) && phdErr.errorCode != pdhInvalidData && phdErr.errorCode != pdhCalcNegativeValue {
		time.Sleep(time.Second)
		farr, err = query.getFormattedCounterArrayDouble(hCounter)
	}
	require.NoError(t, err)
	require.NotEmpty(t, farr)

	rarr, err := query.getRawCounterArray(hCounter)
	require.NoError(t, err)
	require.NotEmpty(t, rarr, "Too")

	require.NoError(t, query.close())
}

func TestWinPerfCountersConfigGet1Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	instances := []string{"_Total"}
	counters := []string{"% Processor Time"}
	perfObjects := []perfObject{{
		ObjectName:    "Processor Information",
		Instances:     instances,
		Counters:      counters,
		Measurement:   "test",
		WarnOnMissing: false,
		FailOnMissing: true,
		IncludeTotal:  false,
	}}

	m := WinPerfCounters{
		PrintValid:    false,
		Object:        perfObjects,
		MaxBufferSize: defaultMaxBufferSize,
		Log:           testutil.Logger{},
		queryCreator:  &performanceQueryCreatorImpl{},
	}

	require.NoError(t, m.parseConfig())
}

func TestWinPerfCountersConfigGet2Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	instances := []string{"_Total"}
	counters := []string{"% Processor Time"}
	perfObjects := []perfObject{{
		ObjectName:    "Processor Information",
		Instances:     instances,
		Counters:      counters,
		Measurement:   "test",
		WarnOnMissing: false,
		FailOnMissing: true,
		IncludeTotal:  false,
	}}

	m := WinPerfCounters{
		PrintValid:    false,
		Object:        perfObjects,
		MaxBufferSize: defaultMaxBufferSize,
		Log:           testutil.Logger{},
		queryCreator:  &performanceQueryCreatorImpl{},
	}

	require.NoError(t, m.parseConfig())

	hostCounters, ok := m.hostCounters["localhost"]
	require.True(t, ok)
	require.Len(t, hostCounters.counters, 1, "There should be exactly one result returned from the counterPath")
}

func TestWinPerfCountersConfigGet3Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	sources := []string{"localhost"}
	instances := []string{"_Total"}
	counters := []string{"% Processor Time", "% Idle Time"}
	perfObjects := []perfObject{{
		Sources:       sources,
		ObjectName:    "Processor Information",
		Instances:     instances,
		Counters:      counters,
		Measurement:   "test",
		WarnOnMissing: false,
		FailOnMissing: true,
		IncludeTotal:  false,
	}}

	m := WinPerfCounters{
		PrintValid:    false,
		Object:        perfObjects,
		MaxBufferSize: defaultMaxBufferSize,
		Log:           testutil.Logger{},
		queryCreator:  &performanceQueryCreatorImpl{},
	}

	require.NoError(t, m.parseConfig())

	hostCounters, ok := m.hostCounters["localhost"]
	require.True(t, ok)
	require.Len(t, hostCounters.counters, 2, "There should be exactly two results returned from the counterPath")
}

func TestWinPerfCountersConfigGet4Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	instances := []string{"_Total", "0,1"}
	counters := []string{"% Processor Time"}
	perfObjects := []perfObject{{
		ObjectName:    "Processor Information",
		Instances:     instances,
		Counters:      counters,
		Measurement:   "test",
		WarnOnMissing: false,
		FailOnMissing: true,
		IncludeTotal:  false,
	}}

	m := WinPerfCounters{
		PrintValid:    false,
		Object:        perfObjects,
		MaxBufferSize: defaultMaxBufferSize,
		Log:           testutil.Logger{},
		queryCreator:  &performanceQueryCreatorImpl{},
	}

	require.NoError(t, m.parseConfig())

	hostCounters, ok := m.hostCounters["localhost"]
	require.True(t, ok)
	require.Len(t, hostCounters.counters, 2, "There should be exactly two results returned from the counterPath")
}

func TestWinPerfCountersConfigGet5Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	instances := []string{"_Total", "0,1"}
	counters := []string{"% Processor Time", "% Idle Time"}
	perfObjects := []perfObject{{
		ObjectName:    "Processor Information",
		Instances:     instances,
		Counters:      counters,
		Measurement:   "test",
		WarnOnMissing: false,
		FailOnMissing: true,
		IncludeTotal:  false,
	}}

	m := WinPerfCounters{
		PrintValid:    false,
		Object:        perfObjects,
		MaxBufferSize: defaultMaxBufferSize,
		Log:           testutil.Logger{},
		queryCreator:  &performanceQueryCreatorImpl{},
	}

	require.NoError(t, m.parseConfig())

	hostCounters, ok := m.hostCounters["localhost"]
	require.True(t, ok)
	require.Len(t, hostCounters.counters, 4, "There should be exactly four results returned from the counterPath")
}

func TestWinPerfCountersConfigGet6Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	instances := []string{"------"}
	counters := []string{"Context Switches/sec"}
	perfObjects := []perfObject{{
		ObjectName:    "System",
		Instances:     instances,
		Counters:      counters,
		Measurement:   "test",
		WarnOnMissing: false,
		FailOnMissing: true,
		IncludeTotal:  false,
	}}

	m := WinPerfCounters{
		PrintValid:    false,
		Object:        perfObjects,
		MaxBufferSize: defaultMaxBufferSize,
		Log:           testutil.Logger{},
		queryCreator:  &performanceQueryCreatorImpl{},
	}

	require.NoError(t, m.parseConfig())

	_, ok := m.hostCounters["localhost"]
	require.True(t, ok)
}

func TestWinPerfCountersConfigGet7Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	instances := []string{"_Total"}
	counters := []string{"% Processor Time", "% Processor TimeERROR", "% Idle Time"}
	perfObjects := []perfObject{{
		ObjectName:  "Processor Information",
		Counters:    counters,
		Instances:   instances,
		Measurement: "test",
	}}

	m := WinPerfCounters{
		PrintValid:    false,
		Object:        perfObjects,
		MaxBufferSize: defaultMaxBufferSize,
		Log:           testutil.Logger{},
		queryCreator:  &performanceQueryCreatorImpl{},
	}

	require.NoError(t, m.parseConfig())

	hostCounters, ok := m.hostCounters["localhost"]
	require.True(t, ok)
	require.Len(t, hostCounters.counters, 2, "There should be exactly two results returned from the counterPath")
}

func TestWinPerfCountersConfigError1Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	instances := []string{"_Total"}
	counters := []string{"% Processor Time"}
	perfObjects := []perfObject{{
		ObjectName:    "Processor InformationERROR",
		Instances:     instances,
		Counters:      counters,
		Measurement:   "test",
		WarnOnMissing: false,
		FailOnMissing: true,
		IncludeTotal:  false,
	}}

	m := WinPerfCounters{
		PrintValid:    false,
		Object:        perfObjects,
		MaxBufferSize: defaultMaxBufferSize,
		Log:           testutil.Logger{},
		queryCreator:  &performanceQueryCreatorImpl{},
	}

	require.Error(t, m.parseConfig())
}

func TestWinPerfCountersConfigError2Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	instances := []string{"SuperERROR"}
	counters := []string{"% C1 Time"}
	perfObjects := []perfObject{{
		ObjectName:    "Processor",
		Instances:     instances,
		Counters:      counters,
		Measurement:   "test",
		WarnOnMissing: false,
		FailOnMissing: true,
		IncludeTotal:  false,
	}}

	m := WinPerfCounters{
		PrintValid:    false,
		Object:        perfObjects,
		MaxBufferSize: defaultMaxBufferSize,
		Log:           testutil.Logger{},
		queryCreator:  &performanceQueryCreatorImpl{},
	}

	require.NoError(t, m.parseConfig())
	var acc testutil.Accumulator
	require.Error(t, m.Gather(&acc))
}

func TestWinPerfCountersConfigError3Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	instances := []string{"_Total"}
	counters := []string{"% Processor TimeERROR"}
	perfObjects := []perfObject{{
		ObjectName:    "Processor Information",
		Instances:     instances,
		Counters:      counters,
		Measurement:   "test",
		WarnOnMissing: false,
		FailOnMissing: true,
		IncludeTotal:  false,
	}}

	m := WinPerfCounters{
		PrintValid:    false,
		Object:        perfObjects,
		MaxBufferSize: defaultMaxBufferSize,
		Log:           testutil.Logger{},
		queryCreator:  &performanceQueryCreatorImpl{},
	}

	require.Error(t, m.parseConfig())
}

func TestWinPerfCountersCollect1Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	instances := []string{"_Total"}
	counters := []string{"Parking Status"}
	perfObjects := []perfObject{{
		ObjectName:    "Processor Information",
		Instances:     instances,
		Counters:      counters,
		Measurement:   "test",
		WarnOnMissing: false,
		FailOnMissing: true,
		IncludeTotal:  false,
	}}

	m := WinPerfCounters{
		PrintValid:    false,
		Object:        perfObjects,
		MaxBufferSize: defaultMaxBufferSize,
		Log:           testutil.Logger{},
		queryCreator:  &performanceQueryCreatorImpl{},
	}

	var acc testutil.Accumulator
	require.NoError(t, m.Gather(&acc))

	time.Sleep(2000 * time.Millisecond)
	require.NoError(t, m.Gather(&acc))
	require.Len(t, acc.Metrics, 2)

	for _, metric := range acc.Metrics {
		_, ok := metric.Fields["Parking_Status"]
		require.True(t, ok)
	}
}

func TestWinPerfCountersCollect2Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	instances := []string{"_Total", "0,0"}
	counters := []string{"Performance Limit Flags"}
	perfObjects := []perfObject{{
		ObjectName:    "Processor Information",
		Instances:     instances,
		Counters:      counters,
		Measurement:   "test",
		WarnOnMissing: false,
		FailOnMissing: true,
		IncludeTotal:  false,
	}}

	m := WinPerfCounters{
		PrintValid:            false,
		UsePerfCounterTime:    true,
		Object:                perfObjects,
		UseWildcardsExpansion: true,
		MaxBufferSize:         defaultMaxBufferSize,
		Log:                   testutil.Logger{},
		queryCreator:          &performanceQueryCreatorImpl{},
	}

	var acc testutil.Accumulator
	require.NoError(t, m.Gather(&acc))

	time.Sleep(2000 * time.Millisecond)
	require.NoError(t, m.Gather(&acc))

	require.Len(t, acc.Metrics, 4)

	for _, metric := range acc.Metrics {
		_, ok := metric.Fields["Performance_Limit_Flags"]
		require.True(t, ok)
	}
}

func TestWinPerfCountersCollectRawIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	instances := []string{"*"}
	counters := []string{"% Idle Time"}
	perfObjects := []perfObject{{
		ObjectName:    "Processor",
		Instances:     instances,
		Counters:      counters,
		Measurement:   "test",
		WarnOnMissing: false,
		FailOnMissing: true,
		IncludeTotal:  false,
		UseRawValues:  true,
	}}

	m := WinPerfCounters{
		PrintValid:            false,
		Object:                perfObjects,
		UseWildcardsExpansion: true,
		MaxBufferSize:         defaultMaxBufferSize,
		Log:                   testutil.Logger{},
		queryCreator:          &performanceQueryCreatorImpl{},
	}
	var acc testutil.Accumulator
	require.NoError(t, m.Gather(&acc))

	time.Sleep(2000 * time.Millisecond)
	require.NoError(t, m.Gather(&acc))
	require.Greater(t, len(acc.Metrics), 1)

	expectedCounter := "Percent_Idle_Time_Raw"
	for _, metric := range acc.Metrics {
		val, ok := metric.Fields[expectedCounter]
		require.True(t, ok, "Expected presence of %s field", expectedCounter)
		valInt64, ok := val.(int64)
		require.Truef(t, ok, "Expected int64, got %T", val)
		require.Positivef(t, valInt64, "Value not positive for metric %#v", metric)
	}

	// Test *Array way
	m = WinPerfCounters{
		PrintValid:            false,
		Object:                perfObjects,
		UseWildcardsExpansion: false,
		MaxBufferSize:         defaultMaxBufferSize,
		Log:                   testutil.Logger{},
		queryCreator:          &performanceQueryCreatorImpl{},
	}
	var acc2 testutil.Accumulator
	require.NoError(t, m.Gather(&acc))

	time.Sleep(2000 * time.Millisecond)
	require.NoError(t, m.Gather(&acc2))
	require.Greater(t, len(acc2.Metrics), 1)

	for _, metric := range acc2.Metrics {
		val, ok := metric.Fields[expectedCounter]
		require.True(t, ok, "Expected presence of %s field", expectedCounter)
		valInt64, ok := val.(int64)
		require.Truef(t, ok, "Expected int64, got %T", val)
		require.Positivef(t, valInt64, "Value not positive for metric %#v", metric)
	}
}
