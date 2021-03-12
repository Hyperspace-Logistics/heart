local app = require('heart.v1')

-- get the value of the cookie 'cookie'
app.get('/cookie', function(ctx)
  return ctx.json({cookie = ctx.cookies('cookie')})
end)

-- set the cookie 'cookie' to the value of the query param 'value'
app.post('/cookie', function(ctx)
  ctx.cookies('cookie', ctx.queryParam('value'))
end)
