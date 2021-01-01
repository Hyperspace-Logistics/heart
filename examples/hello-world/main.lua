app = require('heart.v1')

-- This handler benchmarks at 320k requests a second on my machine :o
app.get('/:name?', function(ctx)
  local name = ctx.pathParams('name')

  if name == '' then
    name = 'world'
  end

  return ctx.json({hello = name})
end)

app.get('/world', function(ctx)
  ctx.redirect('/')
end)
