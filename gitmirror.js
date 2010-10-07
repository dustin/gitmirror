var http = require('http');
var url = require('url');
var path = require('path');
var spawn = require('child_process').spawn;

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

function handleReq(req, res) {
    var preq = url.parse(req.url, true);
    var section = path.normalize(preq.pathname.substring(1));
    if (section[0] == ".") {
        res.writeHead(403, headers);
        res.end("You're doing it wrong\n");
    } else {
        var backgrounded = (preq.query && preq.query.bg == 'false') ? false : true;

        var stdout = [], stderr = [];
        var child = spawn(git, ["remote", "update", "-p"],
                          {'GIT_DIR': path.join(process.cwd(), section)});
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
}

http.createServer(handleReq).listen(8124, "0.0.0.0");
console.log('Server running at http://0.0.0.0:8124/');
