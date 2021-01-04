app = require('heart.v1')
local json = require('heart.v1.json')
local kv = require('heart.v1.kv.disk')

-- create document in bucket
app.post('/documents/:bucket/:id', function(ctx)
  local id = ctx.pathParam('bucket') .. '_documents_' .. ctx.pathParam('id')
  local document = ctx.body().json().document

  if document == nil then
    return ctx.status(400).json({error = 'missing JSON key \'document\''})
  end

  kv.transaction(function(store)
    store.set(id, json.encode(document))
  end)

  ctx.status(201)
  return ''
end)

-- retrieve document in bucket
app.get('/documents/:bucket/:id', function(ctx)
  local id = ctx.pathParam('bucket') .. '_documents_' .. ctx.pathParam('id')
  local document = kv.get(id)

  if document == '' then
    return ctx.status(400).json({error = 'document not found'})
  end

  return ctx.json({document = document})
end)

-- delete document in bucket
app.delete('/documents/:bucket/:id', function(ctx)
  local id = ctx.pathParam('bucket') .. '_documents_' .. ctx.pathParam('id')

  kv.transaction(function(store)
    store.delete(id)
  end)

  return ''
end)

-- list documents in bucket
app.get('/documents', function(ctx)
  return ctx.json({documents = kv.listPairs(10)})
end)
