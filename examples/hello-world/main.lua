app = require('heart.v1')

-- This handler benchmarks at 320k requests a second on my machine :o
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
  return nil
end)
