local core
xpcall(function()
	local exedir = EXEFILE:match('^(.*)/[^/]+$')
	local prefix = exedir:match('^(.*)[^/]+bin$')
	dofile((MACOS_RESOURCES or (prefix and prefix .. '/share/lite-xl' or exedir .. '/data')) .. '/core/start.lua')
	core = require(os.getenv 'LITE_XL_RUNTIME' or 'core')
	core.init()
	core.run()
end, function(err)
	local error_dir
	io.stdout:write('Error: '..tostring(err)..'\n')
	io.stdout:write(debug.traceback(nil, 4)..'\n')
	io.flush()

	if core and core.on_error then
		error_dir = USERDIR
		pcall(core.on_error, err)
	else
		error_dir = system.absolute_path '.'
		local fp = io.open('error.txt', 'wb')
		fp:write('Error: ' .. tostring(err) .. '\n')
		fp:write(debug.traceback(nil, 4)..'\n')
		fp:close()
	end

	system.show_fatal_error('Lite XL internal error',
		'An internal error occurred in a critical part of the application.\n\n'..
		'Please verify the file "error.txt" in the directory '..error_dir)
	os.exit(1)
end)

return core and core.restart_request
