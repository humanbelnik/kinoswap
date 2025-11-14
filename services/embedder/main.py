from server import serve
import os

if __name__ == '__main__':
    delivery_proto = os.getenv('DELIVERY_PROTO', 'GRPC')
    serve(delivery_proto)