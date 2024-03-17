import tkinter as tk
from tkinter import ttk, scrolledtext
import threading
import requests

class DBWorkbenchClient(tk.Tk):
    def __init__(self):
        super().__init__()
        self.title("DB Workbench Client")
        self.geometry("800x600")

        self.tabControl = ttk.Notebook(self)
        
        self.tabGet = ttk.Frame(self.tabControl)
        self.tabSearch = ttk.Frame(self.tabControl)
        self.tabQuery = ttk.Frame(self.tabControl)
        
        self.tabControl.add(self.tabGet, text='GET')
        self.tabControl.add(self.tabSearch, text='SEARCH')
        self.tabControl.add(self.tabQuery, text='QUERY')
        
        self.tabControl.pack(expand=1, fill="both")
        
        self.setupGetTab()
        self.setupSearchTab()
        self.setupQueryTab()

    def setupGetTab(self):
        ttk.Label(self.tabGet, text="ID:").grid(column=0, row=0, padx=10, pady=10)
        self.getIDEntry = ttk.Entry(self.tabGet, width=60)
        self.getIDEntry.grid(column=1, row=0, padx=10, pady=10)
        getButton = ttk.Button(self.tabGet, text="Get", command=self.get)
        getButton.grid(column=2, row=0, padx=10, pady=10)

        self.getResultText = scrolledtext.ScrolledText(self.tabGet, width=70, height=15, wrap=tk.WORD)
        self.getResultText.grid(column=0, row=1, columnspan=3, padx=10, pady=10)

    def setupSearchTab(self):
        ttk.Label(self.tabSearch, text="Keyword:").grid(column=0, row=0, padx=10, pady=10)
        self.searchEntry = ttk.Entry(self.tabSearch, width=60)
        self.searchEntry.grid(column=1, row=0, padx=10, pady=10)
        searchButton = ttk.Button(self.tabSearch, text="Search", command=self.search)
        searchButton.grid(column=2, row=0, padx=10, pady=10)

        self.searchResultText = scrolledtext.ScrolledText(self.tabSearch, width=70, height=15, wrap=tk.WORD)
        self.searchResultText.grid(column=0, row=1, columnspan=3, padx=10, pady=10)

    def setupQueryTab(self):
        ttk.Label(self.tabQuery, text="Terms (comma-separated):").grid(column=0, row=0, padx=10, pady=10)
        self.queryEntry = ttk.Entry(self.tabQuery, width=60)
        self.queryEntry.grid(column=1, row=0, padx=10, pady=10)
        queryButton = ttk.Button(self.tabQuery, text="Query", command=self.query)
        queryButton.grid(column=2, row=0, padx=10, pady=10)

        self.queryResultText = scrolledtext.ScrolledText(self.tabQuery, width=70, height=15, wrap=tk.WORD)
        self.queryResultText.grid(column=0, row=1, columnspan=3, padx=10, pady=10)

    def db_request(self, endpoint, method="get", params=None):
        try:
            if method == "get":
                response = requests.get(endpoint, params=params)
            # Add other methods as needed
            
            if response.status_code == 200:
                return response.json()
            else:
                return {"error": f"Server returned status code {response.status_code}"}
        except Exception as e:
            return {"error": str(e)}

    def get(self):
        id = self.getIDEntry.get()
        endpoint = f"http://localhost:8080/get/{id}"
        threading.Thread(target=self.execute_db_request, args=(endpoint, self.getResultText)).start()

    def search(self):
        keyword = self.searchEntry.get()
        endpoint = f"http://localhost:8080/search/{keyword}"
        threading.Thread(target=self.execute_db_request, args=(endpoint, self.searchResultText)).start()

    def query(self):
        terms = self.queryEntry.get().split(',')
        endpoint = f"http://localhost:8080/query/{'/'.join(terms)}"
        threading.Thread(target=self.execute_db_request, args=(endpoint, self.queryResultText)).start()

    def execute_db_request(self, endpoint, result_widget):
        response = self.db_request(endpoint)
        result_widget.delete('1.0', tk.END)
        result_widget.insert(tk.INSERT, str(response))

if __name__ == "__main__":
    app = DBWorkbenchClient()
    app.mainloop()
