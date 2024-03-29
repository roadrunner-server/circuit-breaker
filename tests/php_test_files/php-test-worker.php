<?php
/**
 * @var Goridge\RelayInterface $relay
 */

use Nyholm\Psr7\Factory\Psr17Factory;
use Nyholm\Psr7\Response;
use Spiral\Goridge;
use Spiral\RoadRunner;

ini_set('display_errors', 'stderr');
require __DIR__ . "/vendor/autoload.php";

$worker = RoadRunner\Worker::create();
$psr7 = new RoadRunner\Http\PSR7Worker(
    $worker,
    new Psr17Factory(),
    new Psr17Factory(),
    new Psr17Factory()
);

while ($req = $psr7->waitRequest()) {
    try {
        $resp = new Response();

        if ($req->getUri()->getPath() === "/error") {
            $resp = $resp->withStatus(500);
        }
        $resp->getBody()->write("hello world");

        $psr7->respond($resp);
    } catch (Throwable $e) {
        $psr7->getWorker()->error((string)$e);
    }
}
