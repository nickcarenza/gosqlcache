gosqlcache
==========
 > A postgres caching driver on top of lib/pq that implements database/sql

How it works
------------
Gosqlcache registers itself as a driver and allows you to register query strings to be cached for a period of time. When you run a registered query, the driver will match the query and argument list against cached results. If a registered query is run and there are no cached results, it will cache them for you.

Exec statements and unregistered queries pass directly through to lib/pq.

Warnings
--------
 - Might have unexpected results when using function calls or dynamic values in query strings. i.e. now()

TODOs
-----
 - synchronize multiple queries to the same result set and only run it once
 - allow multiple sql.DB to have different caches
 - add stats