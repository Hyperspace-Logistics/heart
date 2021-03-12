-- _heart holds all of the routing state for the app
-- it's a global so it can be used anywhere in the app without being passed around as a single variable
-- it also makes lookup easier
_heart = {routes = {}, ctx = require('heart.v1.context')}

function registerCallback(method, path, callback)
  if _heart.routes[path] == nil then
    _heart.routes[path] = {}
  end

  _heart.routes[path][method] = callback
end

function _heart.get(path, callback)
  registerCallback('get', path, callback)
end

function _heart.head(path, callback)
  registerCallback('head', path, callback)
end

function _heart.post(path, callback)
  registerCallback('post', path, callback)
end

function _heart.put(path, callback)
  registerCallback('put', path, callback)
end

function _heart.delete(path, callback)
  registerCallback('delete', path, callback)
end

function _heart.options(path, callback)
  registerCallback('options', path, callback)
end

function _heart.trace(path, callback)
  registerCallback('trace', path, callback)
end

function _heart.patch(path, callback)
  registerCallback('patch', path, callback)
end

function _heart.static(route, filepath)
  _static(route, filepath)
end

function _heart.notfound(callback)
  registerCallback('_not_found', '/', callback)
end

package.preload['heart.v1'] = function()
  return _heart
end
