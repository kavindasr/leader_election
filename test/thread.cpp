#include <iostream>
#include <thread>

// Function for the first thread
void threadFunction1() {
    for (int i = 0; i < 5; ++i) {
        std::cout << "Thread 1: " << i << std::endl;
    }
}

// Function for the second thread
void threadFunction2() {
    for (int i = 0; i < 5; ++i) {
        std::cout << "Thread 2: " << i << std::endl;
    }
}

int main() {
    // Create two thread objects
    std::thread t1(threadFunction1);
    std::thread t2(threadFunction2);

    // Wait for both threads to finish
    t1.join();
    t2.join();

    return 0;
}
