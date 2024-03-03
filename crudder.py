import argparse
import requests
import re

def make_request(method, url, output_file=None):
    try:
        response = requests.request(method, url)
        # Prepare the result string with a regex to remove the specified pattern
        result = re.sub(r'request to .*:', '', f"{method} request to {url}: {response.status_code}")
        result = f"{method} {result}\n"  # Reformatted result
        print(result, end='')
        if output_file:
            with open(output_file, 'a') as file:
                file.write(result)
    except Exception as e:
        error_msg = f"Failed to make {method} request to {url}: {str(e)}\n"
        error_msg = re.sub(r'request to .*:', '', error_msg)
        print(error_msg, end='')
        if output_file:
            with open(output_file, 'a') as file:
                file.write(error_msg)

def read_file(file_path):
    with open(file_path, 'r') as file:
        return [line.strip() for line in file.readlines()]

def main():
    parser = argparse.ArgumentParser(description='Make HTTP requests to URLs listed in a file and optionally log results.')
    parser.add_argument('-m', '--methods', type=str, help='Methods to test: c (POST), r (GET), u (PUT), d (DELETE)')
    parser.add_argument('urls_file', type=str, help='File containing URLs, one per line')
    parser.add_argument('-o', '--output', type=str, help='Optional: Output file to write the results', default=None)
    args = parser.parse_args()

    method_map = {'u': 'PUT', 'r': 'GET', 'd': 'DELETE', 'c': 'POST'}
    selected_methods = [method_map[method] for method in args.methods if method in method_map]

    if not selected_methods:
        print("No valid methods selected. Please use -m with 'c', 'r', 'u', or 'd'.")
        return

    urls = read_file(args.urls_file)

    if args.output:
        # Clear the output file content before appending new results
        open(args.output, 'w').close()

    for url in urls:
        if args.output:
            with open(args.output, 'a') as file:
                file.write(f"Endpoint: {url}\n")
        print(f"Endpoint: {url}")
        for method in selected_methods:
            make_request(method, url, args.output)

if __name__ == "__main__":
    main()
