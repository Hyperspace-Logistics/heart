app = require('heart')

app.get('/:name?', function(ctx)
  local name = ctx.pathParams('name')

  if name == '' then
    return 'Hello, world!'
  else
    return 'Hello, ' .. name .. '!'
  end
end)

app.get('/world', function(ctx)
  ctx.redirect('/')
end)
