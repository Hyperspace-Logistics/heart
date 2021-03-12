local app = require('heart.v1')
local kv = require('heart.v1.kv.memory')

local function get_hits()
  local hits = ''
  kv.serialTransaction(function(store)
    hits = store.get('hits')
    if hits == '' then
      hits = '0'
    end
    hits = tonumber(hits)

    store.set('hits', tostring(hits + 1))
  end)

  return hits
end

app.get('/', function(ctx)
  return ctx.json({hits = get_hits()})
end)
