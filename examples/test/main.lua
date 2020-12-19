app = require('heart.v1')

app.get('/', function(ctx)
  return ctx.json({hello = 'world'})
end)
