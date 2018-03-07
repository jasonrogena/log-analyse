package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
)

var (
	writeCaches           map[string]*writeCache
	writeCacheRequestChan chan writeCacheRequest
	writeCacheWaitGroup   *sync.WaitGroup
)

type writeCache struct {
	db           *sql.DB
	table        string
	queries      map[string][][]interface{}
	dependencies []*writeCache
	cacheCount   int64
	maxCacheSize int64
}

type writeCacheRequest struct {
	table string
	query string
	args  []interface{}
}

func openCache() {
	writeCacheRequestChan = make(chan writeCacheRequest, 1000)
	writeCacheWaitGroup = new(sync.WaitGroup)
	writeCacheWaitGroup.Add(1)
	go addToWriteCache(writeCacheRequestChan)
}

func closeCache() {
	close(writeCacheRequestChan)
	writeCacheWaitGroup.Wait()
}

func createWriteCache(db *sql.DB, tableName string, maxCacheSixe int64, dependencyTables []string) {
	var deps []*writeCache
	for _, curDepName := range dependencyTables {
		deps = append(deps, writeCaches[curDepName])
	}
	tableWriteCache := writeCache{
		db:           db,
		table:        tableName,
		queries:      make(map[string][][]interface{}),
		dependencies: deps,
		cacheCount:   0,
		maxCacheSize: maxCacheSixe}
	writeCaches[tableName] = &tableWriteCache
}

func addToWriteCache(cacheRequests <-chan writeCacheRequest) {
	defer writeCacheWaitGroup.Done()
	defer unloadAllWriteCaches()

	for request := range cacheRequests {
		if tableWriteCache, writeCacheOk := writeCaches[request.table]; writeCacheOk {
			// Check if query exists in cache
			if _, queryInCache := tableWriteCache.queries[request.query]; !queryInCache {
				tableWriteCache.queries[request.query] = make([][]interface{}, 0)
			}
			tableWriteCache.queries[request.query] = append(tableWriteCache.queries[request.query], request.args)
			tableWriteCache.cacheCount++
			if tableWriteCache.cacheCount > tableWriteCache.maxCacheSize {
				unloadWriteCache(tableWriteCache)
			}
		}
	}
}

func unloadAllWriteCaches() {
	for _, curWriteCache := range writeCaches {
		unloadWriteCache(curWriteCache)
	}
}

func unloadWriteCache(cache *writeCache) {
	// Make sure all the dependencies are unloaded first
	for _, curDepCache := range cache.dependencies {
		unloadWriteCache(curDepCache)
	}

	for query, args := range cache.queries {
		if len(args) > 0 {
			allArgs := args[0]

			argsSetString := "(?" + strings.Repeat(", ?", len(args[0])-1) + ")"
			query = query + strings.Repeat(", "+argsSetString, len(args)-1)
			for i := 1; i < len(args); i++ {
				allArgs = append(allArgs, args[i]...)
			}

			_, insertErr := runInsertQuery(cache.db, query, allArgs, 0)
			if insertErr != nil {
				fmt.Fprintln(os.Stderr, insertErr.Error())
			}
		}
	}

	cache.queries = make(map[string][][]interface{})
	cache.cacheCount = 0
}

func cacheInsert(db *sql.DB, table string, query string, args []interface{}, idArg int) (string, error) {
	if _, writeCacheOk := writeCaches[table]; writeCacheOk {
		writeCacheRequestChan <- writeCacheRequest{table: table, query: query, args: args}
		cachedUUID, uuidOk := args[idArg].(string)
		if uuidOk {
			return cachedUUID, nil
		}

		return "", errors.New("Could not get the cached UUID")
	}
	return runInsertQuery(db, query, args, idArg)
}
