# Heart 💜

A high performance Lua web framework with a simple, powerful API

## Overview

Heart combines Go's [fasthttp](https://github.com/valyala/fasthttp) with [LuaJIT](https://luajit.org/)
to create an insanely fast Lua web server.

It also comes with a performant key-value store API backed by [BadgerDB](https://github.com/dgraph-io/badger)
that can store data both in memory and on disk.

## Features

- Single binary
- High throughput
- Low latency
- Fast K/V store
- Versioned API
- Stable Lua 5.1 VM

## Caveats

Global state, like with any parallel web server, is highly discouraged. For performance reasons Heart keeps a
pool of Lua state to reuse in subsequent requests. This can make global state seem stable at low request volumes
but it's functionally random under load. Be cautious!
