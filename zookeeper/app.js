const zookeeper = require('node-zookeeper-client');

const client = zookeeper.createClient('localhost:2181');
client.connect();

client.once('connected', () => {
    console.log('Connected to ZooKeeper');
});