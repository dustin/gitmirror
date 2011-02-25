var http = require('http');
var url = require('url');
var path = require('path');
var spawn = require('child_process').spawn;
var querystring = require('querystring');
var fs = require('fs');

// 0 == node, 1 == [this script]
var git = process.argv[2];
var base = process.argv[3];

process.chdir(base);

var headers = {
    'Content-Type': 'text/plain',
};

function labeledArray(l, a) {
    return "---- " + l + " ----\n" + a.join('') + "\n----\n";
}

function getFullPath(section) {
    return path.join(process.cwd(), section);
}

function logCommand(name, child) {
    var stdout = [], stderr = [];
    child.stdout.on('data', function (data) { stdout.push(data); });
    child.stderr.on('data', function (data) { stderr.push(data); });
    child.on('exit', function (code) {
                 console.log('Completed execution of ' + name + ': ' + code);
                 if (code != 0) {
                     console.log(labeledArray('stderr', stderr));
                     console.log(labeledArray('stdout', stdout));
                     }
                 });
}

function spawnIfExists(cb, path, cmd, args, stdin, cmdOpts) {
    fs.stat(path, function(err, stats) {
        if (err) {
            cb();
        } else {
            console.log("Running " + cmd);
            var child = spawn(cmd, args, cmdOpts);
            if (stdin) {
                child.stdin.write(stdin);
                child.stdin.end();
            }
            logCommand(cmd, child);
            child.on('exit', cb);
        }
    });
}

function runSequence(functions) {
    var current = 0;
    function theCallback() {
        var prev = current;
        if (prev >= functions.length) {
            return;
        }
        ++current;
        functions[prev](theCallback);
    }
    theCallback();
}

function runHooks(section, payload) {
    var thePath = getFullPath(section);
    var payloadString = payload ? JSON.stringify(payload) : "";

    var functions = [
        function(cb) {
            spawnIfExists(cb, thePath, git, ['gc', '--auto'],
                undefined, {'cwd': thePath});
        },
        function(cb) {
            var fullPath = path.join(thePath, 'hooks/post-fetch');
            spawnIfExists(cb, fullPath, fullPath, [],
                          payloadString, {'cwd': thePath});
        },
        function(cb) {
            var fullPath = path.resolve('bin/post-fetch');
            spawnIfExists(cb, fullPath, fullPath, [],
                          payloadString, {'cwd': thePath});
        }
    ];

    runSequence(functions);
}

function handleComplete(res, child, section, backgrounded, payload) {
    var stdout = [], stderr = [];
    child.stdout.on('data', function (data) { stdout.push(data); });
    child.stderr.on('data', function (data) { stderr.push(data); });
    if (backgrounded) {
        res.writeHead(202, headers);
        res.end();
        child.on('exit', function (code) {
                     console.log('bg child process updating ' + section
                                 + ' completed with exit code ' + code);
                     if (code != 0) {
                         console.log(labeledArray('stderr', stderr));
                         console.log(labeledArray('stdout', stdout));
                     }
                 });
    } else {
        child.on('exit', function (code) {
                     console.log('fg child process updating ' + section
                                 + ' completed with exit code ' + code);
                     res.writeHead(code == 0 ? 200 : 500, headers);
                     res.end(labeledArray('stderr', stderr)
                             + labeledArray('stdout', stdout));
                 });
    }
    child.on('exit', function(code) { runHooks(section, payload); });
}

function gitUpdate(res, section, backgrounded, payload) {
    var child = spawn(git, ["remote", "update", "-p"],
			         {'cwd': getFullPath(section)});
    handleComplete(res, child, section, backgrounded, payload);
}

function createRepo(res, section, backgrounded, payload) {
    var repo = "git://github.com/" + payload.repository.owner.name +
        "/" + payload.repository.name + ".git";
    if (payload.repository.private) {
        repo = "git@github.com:" + payload.repository.owner.name +
            "/" + payload.repository.name + ".git";
    }
    var child = spawn(git, ["clone", "--mirror", "--bare", repo, getFullPath(section)]);
    handleComplete(res, child, section, backgrounded, payload);
}

function githubPost(res, section, backgrounded, payload) {
    console.log("Processing " + payload.repository.owner.name + "/"
                + payload.repository.name);
    fs.stat(getFullPath(section), function(err, stats) {
                if (err || !stats.isDirectory()) {
                    createRepo(res, section, backgrounded, payload);
                } else {
                    gitUpdate(res, section, backgrounded, payload);
                }
            });
}

function handleReq(req, res) {
    var preq = url.parse(req.url, true);
    var section = path.normalize(preq.pathname.substring(1));
    if (section[0] == ".") {
        res.writeHead(403, headers);
        res.end("You're doing it wrong\n");
    } else {
        var backgrounded = (preq.query && preq.query.bg == 'false') ? false : true;
        if (req.method === "GET") {
            gitUpdate(res, section, backgrounded);
        } else if (req.method === "POST") {
            req.content = '';
            req.addListener("data", function(chunk) { req.content += chunk; });
            req.addListener("end", function() {
                                var query = querystring.parse(req.content);
                                if (query.payload) {
                                    githubPost(res, section, backgrounded,
                                               JSON.parse(query.payload));
                                } else {
                                    gitUpdate(res, section, backgrounded);
                                }
	                        });
        }
    }
}

http.createServer(handleReq).listen(8124, "0.0.0.0");
console.log('Server running at http://0.0.0.0:8124/');
