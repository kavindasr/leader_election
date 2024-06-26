#include <iostream>
#include <mpi.h>
#include <thread>
#include <vector>
#include <chrono>

using namespace std;

int initialNodeConfig[9][3] = {
        {2, 3, 345},
        {12, 45, 234},
        {35, 45, 533},
        {1, 38, 234},
        {4, 5, 50},
        {50, 3, 98},
        {22, 2, 144},
        {34, 33, 233},
        {11, 39, 235}
};
int x, y, elevel, leader;

// Function to be executed by each thread
void count(int argc, char** argv) {
    int rank;
    // Initialize MPI environment
    MPI_Init_thread(&argc, &argv, MPI_THREAD_SERIALIZED, nullptr);
    MPI_Comm_rank(MPI_COMM_WORLD, &rank);

    x = initialNodeConfig[rank][0];
    y = initialNodeConfig[rank][1];
    elevel = initialNodeConfig[rank][2];
    cout<< "Running thread: " << rank << endl;
    while(elevel > 0) {
        // cout << "Thread: " << rank << " energy level is: " << elevel<< endl;
        // Perform some computation here (e.g., use rank for different tasks)
        // Sleep 1 second
        this_thread::sleep_for(chrono::milliseconds(1000));
        elevel --;
    }
}

void threadFunction2(int argc, char** argv) {
    // Initialize MPI environment
    MPI_Init_thread(&argc, &argv, MPI_THREAD_SERIALIZED, nullptr);

    int rank, size;
    MPI_Comm_rank(MPI_COMM_WORLD, &rank); // Get the process ID (rank)
    MPI_Comm_size(MPI_COMM_WORLD, &size); // Get the total number of processes

    int data = 0;
    if (rank == leader) {
        cout << "Leader: " << rank << endl;
        data = 100;
    }
    // Move broadcast block to separate function becuase this is run after completing the t1 and t2

    // Broadcast the data from root (rank 0) to all other processes
    MPI_Bcast(&data, 1, MPI_INT, 0, MPI_COMM_WORLD);

    cout << "Rank " << rank << " received data " << data << endl;
    
    // Finalize MPI environment
    MPI_Finalize();
}


int main(int argc, char** argv) {

    leader = 0; // To test broadcast

    thread t1(count, argc, argv);
    thread t2(threadFunction2, argc, argv);

    t1.join();
    t2.join();
    
    return 0;
}
