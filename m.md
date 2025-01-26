To make the transactions truly distributed, we need several key modifications to the architecture. Here are the necessary steps:

Implement Two-Phase Commit (2PC)
First, modify the Transaction struct to support 2PC:

Add Prepare Phase
Add prepare phase to Transaction.Commit:

Add Network Communication Layer
Create a new file server/network/coordinator.go:

Add Participant API Endpoints
Modify routes/routes.go:

Add Node Configuration
Create server/config/nodes.go:

Update Docker Configuration
Modify docker-compose.yml:

Add Transaction Recovery
Add to models/transaction.go:

These changes will:

Implement proper 2PC protocol
Add network communication between nodes
Support multiple database nodes
Handle node failures and recovery
Maintain transaction consistency across nodes
The system will now be truly distributed with:

Multiple participating nodes
Coordinated transaction commits
Failure recovery
Network communication
Distributed state management
Update the client to be aware of the distributed nature:
// In client/src/services/api.js
export const userService = {
    async deposit(accountData) {
        const response = await fetch(`${API_URL}/accounts`, {
            method: 'PATCH',
            headers: {
                'Content-Type': 'application/json',
                'X-Transaction-ID': crypto.randomUUID(), // Add transaction tracking
            },
            body: JSON.stringify(accountData),
        });
        
        // Check for distributed transaction status
        if (response.headers.get('X-Transaction-Status') === 'preparing') {
            return this.pollTransactionStatus(response.headers.get('X-Transaction-ID'));
        }
        
        return response.json();
    }
}



A concurrent application distributed at the student's choice. The application must comply with the following requirements:
- to be distributed, but not simply, client-server, but on several levels (client/web - business/middleware - data, etc.).
- to involve aspects of competition at the level of manipulated external data (i.e. transactions in databases).
- to use two different databases (at least 3 tables) and to use distributed transactions. (it is not mandatory to have 2 distinct database servers)
- to have at least 6-8 operations/use cases.
- Very important: To implement a distributed transaction at the application level. That is, you will consider that a transaction does not consist of read() and write() operations of memory pages as we do in the course, but you will consider a transaction as consisting of simple SQL operations (minimum 3 SQL instructions - insert, delete, update , select) of course, these SQL operations will operate on different tables. You must ensure, at the application level, the ACID properties of this transaction. In other words, implement the following:
a planning algorithm (i.e. concurrency control algorithm) from those discussed in the course (based on blocks or orders, timestamps, etc.); the application should use 2 databases. Those who implement planning based on the ordering of timestamps must also implement a multiversioning mechanism and automatically restart any transaction that the planner aborts.
a rollback mechanism discussed in the course (multiversions, rollback for each simple SQL statement, etc.)
a commit mechanism (can be thought together with rollback cell)
a mechanism for detecting and solving deadlocks (graphs/conflict lists, etc.)
- Attention:: The focus of the application must fall on the implementation of the transactional system, not on use cases, the web interface or frameworks you have used. You can use frameworks that make your work easier (e.g. Hibernate or other JPA, Spring, .NET MVC, etc.), but you must not use any kind of transactional support from them.

Fa la tranzactie intre cont a si b sa stea o perioada sa ramana tranzactia activa,
sa pot sa fac teste pe ea.

Fa si niste frontend. Fa celalte cazuri de utilizare putin mai complexe (combina maybe login la user)