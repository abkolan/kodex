const zookeeper = require('node-zookeeper-client');
const path = require('path');

// Configuration
const ZOOKEEPER_CONNECTION = 'localhost:2181'; // Change to your ZooKeeper server
const ELECTION_PATH = '/election';
const NODE_PREFIX = 'candidate_';

// Create a client
const client = zookeeper.createClient(ZOOKEEPER_CONNECTION);
let myNodePath = null;
let leaderId = null;
let isLeader = false;

// Connect to ZooKeeper
client.connect();

client.once('connected', async () => {
  console.log('Connected to ZooKeeper server');
  
  // Ensure the election path exists
  try {
    await createElectionPath();
    
    // Create an ephemeral sequential node for this instance
    myNodePath = await createCandidateNode();
    console.log(`Created candidate node: ${myNodePath}`);
    
    // Participate in the election
    await electLeader();
  } catch (error) {
    console.error('Error in leader election:', error);
  }
});

// Create the election znode if it doesn't exist
function createElectionPath() {
  return new Promise((resolve, reject) => {
    client.exists(ELECTION_PATH, (error, stat) => {
      if (error) {
        return reject(error);
      }
      
      if (stat) {
        return resolve(); // Path already exists
      }
      
      // Create the parent path
      client.create(ELECTION_PATH, null, zookeeper.ACL.OPEN_ACL_UNSAFE, (createError) => {
        if (createError) {
          return reject(createError);
        }
        resolve();
      });
    });
  });
}

// Create a sequential ephemeral node
function createCandidateNode() {
  return new Promise((resolve, reject) => {
    const candidatePath = path.join(ELECTION_PATH, NODE_PREFIX);
    
    // Create an ephemeral sequential node
    client.create(
      candidatePath,
      Buffer.from(process.pid.toString()),
      zookeeper.ACL.OPEN_ACL_UNSAFE,
      zookeeper.CreateMode.EPHEMERAL_SEQUENTIAL,
      (error, path) => {
        if (error) {
          return reject(error);
        }
        resolve(path);
      }
    );
  });
}

// Elect a leader
function electLeader() {
  return new Promise((resolve, reject) => {
    // Get all children of the election path
    client.getChildren(ELECTION_PATH, watchElection, (error, children) => {
      if (error) {
        return reject(error);
      }
      
      // Sort the children (sequential nodes)
      children.sort();
      
      // Get the node with the smallest sequence number
      const leader = children[0];
      const leaderId = path.join(ELECTION_PATH, leader);
      
      // Check if current node is the leader
      const myNodeName = path.basename(myNodePath);
      if (leader === myNodeName) {
        if (!isLeader) {
          isLeader = true;
          console.log('I am the leader now!');
          
          // Do leader-specific tasks here
          startLeaderTasks();
        }
      } else {
        isLeader = false;
        console.log(`I am a follower. Leader is ${leader}`);
        
        // Watch the node before me in the sequence
        const myIndex = children.indexOf(myNodeName);
        if (myIndex > 0) {
          const watchPath = path.join(ELECTION_PATH, children[myIndex - 1]);
          console.log(`Watching node ${watchPath}`);
          
          // Set a watch on the previous node
          client.exists(watchPath, watchPreviousNode, (error, stat) => {
            if (error) {
              return reject(error);
            }
          });
        }
      }
      
      resolve();
    });
  });
}

// Watch for changes in the election path (new candidates joining or leaving)
function watchElection(event) {
  console.log(`Election event: ${event.type}`);
  electLeader().catch(console.error);
}

// Watch the previous node for failure
function watchPreviousNode(event) {
  console.log(`Previous node event: ${event.type}`);
  if (event.type === zookeeper.Event.NODE_DELETED) {
    console.log('Node before me is gone, re-running election');
    electLeader().catch(console.error);
  }
}

// Leader tasks
function startLeaderTasks() {
  // This is where the leader performs its specific duties
  console.log('Starting leader tasks...');
  
  // Example: update a heartbeat every 5 seconds
  const heartbeatInterval = setInterval(() => {
    if (!isLeader) {
      clearInterval(heartbeatInterval);
      return;
    }
    
    const timestamp = new Date().toISOString();
    console.log(`Leader heartbeat: ${timestamp}`);
    
    // Update a node with the heartbeat
    const heartbeatPath = '/heartbeat';
    client.exists(heartbeatPath, (error, stat) => {
      if (error) {
        console.error('Error checking heartbeat path:', error);
        return;
      }
      
      if (stat) {
        // Update existing node
        client.setData(heartbeatPath, Buffer.from(timestamp), -1, (error) => {
          if (error) console.error('Error updating heartbeat:', error);
        });
      } else {
        // Create new node
        client.create(
          heartbeatPath,
          Buffer.from(timestamp),
          zookeeper.ACL.OPEN_ACL_UNSAFE,
          (error) => {
            if (error) console.error('Error creating heartbeat:', error);
          }
        );
      }
    });
  }, 5000);
}

// Handle process termination
process.on('SIGINT', () => {
  console.log('Closing connection to ZooKeeper');
  client.close();
  process.exit(0);
});