var http = require('http');
var url = require('url');
var path = require('path');
var exec = require('child_process').exec;
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

function handleComplete(res, child, section, backgrounded) {
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
}

function gitUpdate(res, section, backgrounded) {
    var child = exec(git + " remote update -p && " + git + " gc --auto"
                     + " && test -x hooks/post-fetch && ./hooks/post-fetch",
			         {'cwd': getFullPath(section)});
    handleComplete(res, child, section, backgrounded);
}

function createRepo(res, section, backgrounded, payload) {
    var repo = "git://github.com/" + payload.repository.owner.name +
        "/" + payload.repository.name + ".git";
    if (payload.repository.private) {
        repo = "git@github.com:" + payload.repository.owner.name +
            "/" + payload.repository.name + ".git";
    }
    var child = exec(git + " clone --mirror --bare " + repo + " " + getFullPath(section));
    handleComplete(res, child, section, backgrounded);
}

function githubPost(res, section, backgrounded, payload) {
    console.log("Processing " + payload.repository.owner.name + "/"
                + payload.repository.name);
    fs.stat(getFullPath(section), function(err, stats) {
                if (err || !stats.isDirectory()) {
                    createRepo(res, section,
                    backgrounded,
                    payload);
                } else {
                    gitUpdate(res, section,
                    backgrounded);
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
