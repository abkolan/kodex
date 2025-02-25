const zookeeper = require('node-zookeeper-client');

// Connect to ZooKeeper
const client = zookeeper.createClient('localhost:2181');
client.connect();

client.once('connected', () => {
    console.log('✅ Connected to ZooKeeper');

    const parentPath = '/config';
    const configPath = '/config/db-url';

    // Ensure parent node exists
    client.exists(parentPath, (err, stat) => {
        if (err) return console.error(err);
        if (!stat) {
            console.log(`⚠️ Parent node ${parentPath} does not exist. Creating it...`);
            client.create(parentPath, Buffer.from(''), zookeeper.CreateMode.PERSISTENT, (error) => {
                if (!error) console.log(`✅ Created parent node ${parentPath}`);
            });
        }
    });

    // Ensure config node exists
    client.exists(configPath, (err, stat) => {
        if (err) return console.error(err);
        if (!stat) {
            console.log(`⚠️ Config node ${configPath} does not exist. Creating it...`);
            client.create(configPath, Buffer.from('mongodb://localhost:27017'), zookeeper.CreateMode.PERSISTENT, (error) => {
                if (!error) console.log('✅ Configuration stored successfully!');
            });
        }
    });

    // Function to read and log config value
    const readConfig = () => {
        client.getData(configPath, (event) => {
            console.log('⚡ Configuration updated:', event);
            readConfig(); // Re-watch after change
        }, (error, data) => {
            if (error) {
                if (error.code === -101) {
                    console.log(`❌ Node ${configPath} does not exist.`);
                } else {
                    console.error(error);
                }
                return;
            }
            console.log(`📌 Current Config Value: ${data.toString()}`);
        });
    };

    // Start watching for changes
    readConfig();
});

// Keep process running indefinitely
process.stdin.resume();