from locust import HttpUser, task, between, events
import json
import logging
import threading

class TransactionUser(HttpUser):
    host = "http://localhost:8080"
    wait_time = between(0.1, 0.5)
    accounts = []
    user_count = 0
    count_lock = threading.Lock()
    
    def get_next_user_number(self):
        with self.count_lock:
            current = TransactionUser.user_count
            TransactionUser.user_count += 1
            return current
    
    def on_start(self):
        try:
            user_num1 = self.get_next_user_number()
            user_num2 = self.get_next_user_number()
            
            user1_id = self.create_user(f"user{user_num1}")
            user2_id = self.create_user(f"user{user_num2}")
            
            logging.info(f"Created users: user{user_num1}, user{user_num2}")
            
            if not user1_id or not user2_id:
                logging.error("Failed to create users")
                return

            account1_id = self.create_account(user1_id)
            account2_id = self.create_account(user2_id)
            if not account1_id or not account2_id:
                logging.error("Failed to create accounts")
                return

            self.accounts = [account1_id, account2_id]

            if not self.deposit(account1_id, 100):
                logging.error(f"Failed to deposit to account {account1_id}")
            if not self.deposit(account2_id, 100):
                logging.error(f"Failed to deposit to account {account2_id}")

        except Exception as e:
            logging.error(f"Error in on_start: {e}")
    
    
    def create_user(self, username):
        try:
            response = self.client.post("/users", 
                json={"username": username})
            if response.status_code == 201:
                return response.json()["id"]
            logging.error(f"Create user failed: {response.text}")
            return None
        except Exception as e:
            logging.error(f"Create user error: {str(e)}")
            return None

    def create_account(self, user_id):
        try:
            response = self.client.post("/accounts", 
                json={"user_id": user_id})
            if response.status_code == 201:
                return response.json()["id"]
            logging.error(f"Create account failed: {response.text}")
            return None
        except Exception as e:
            logging.error(f"Create account error: {str(e)}")
            return None

    def deposit(self, account_id, amount):
        try:
            response = self.client.patch("/accounts",
                json={"account_id": account_id,"amount": amount})
            return response.status_code == 200
        except Exception as e:
            logging.error(f"Deposit error: {str(e)}")
            return False

    @task
    def transfer_1_to_2(self):
        if len(self.accounts) != 2:
            return
        try:
            response = self.client.post("/accounts/transfer", 
                json={
                    "from_account_id": self.accounts[0],
                    "to_account_id": self.accounts[1],
                    "amount": 5
                })
            events.request.fire(
                request_type="transfer",
                name="1->2",
                response_time=response.elapsed.total_seconds() * 1000,
                response_length=len(response.text),
                exception=None if response.status_code == 200 else response.text
            )
        except Exception as e:
            logging.error(f"Transfer 1->2 error: {str(e)}")

    @task
    def transfer_2_to_1(self):
        if len(self.accounts) != 2:
            return
        try:
            response = self.client.post("/accounts/transfer", 
                json={
                    "from_account_id": self.accounts[1],
                    "to_account_id": self.accounts[0],
                    "amount": 5
                })
            events.request.fire(
                request_type="transfer",
                name="2->1",
                response_time=response.elapsed.total_seconds() * 1000,
                response_length=len(response.text),
                exception=None if response.status_code == 200 else response.text
            )
        except Exception as e:
            logging.error(f"Transfer 2->1 error: {str(e)}")