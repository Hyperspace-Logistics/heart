app = require('heart.v1')

-- This handler benchmarks at 270k requests a second on my machine :o
-- with an average latency of 183 microseconds
-- max request latency is 18 milliseconds and I suspect most of that is GC pause
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
