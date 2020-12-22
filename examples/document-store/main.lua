app = require('heart.v1')
local json = require('heart.v1.json')
local kv = require('heart.v1.kv.disk')

-- create new document
app.post('/documents/:id', function(ctx)
  local id = 'document_' .. ctx.pathParams('id')
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

-- read existing document
app.get('/documents/:id', function(ctx)
  local id = 'document_' .. ctx.pathParams('id')
  local document = kv.get(id)

  if document == '' then
    return ctx.status(400).json({error = 'document not found'})
  end

  return ctx.json({document = document})
end)
