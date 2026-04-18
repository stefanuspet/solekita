import 'package:flutter_dotenv/flutter_dotenv.dart';

class ApiConstants {
  static final baseUrl = dotenv.env['API_URL'];

  // AUTH
  static const login = '/auth/login';
  static const register = '/auth/register';
  static const refresh = '/auth/refresh';
  static const logout = '/auth/logout';

  // OUTLETS
  static const outletMe = '/outlets/me';
  static const outletSummary = '/outlets/me/summary';
  static const outletStats = '/outlets/me/stats';

  // USERS
  static const users = '/users';

  // TREATMENTS
  static const treatments = '/treatments';

  // CUSTOMERS
  static const customers = '/customers';

  // ORDERS
  static const orders = '/orders';
  static const orderById = '/orders/:id';
  static const orderStatus = '/orders/:id/status';
  static const orderCancel = '/orders/:id/cancel';
  static const orderPrice = '/orders/:id/price';
  static const orderPhotos = '/orders/:id/photos';

  // ATTENDANCES
  static const attendances = '/attendances';
  static const todayAttendances = '/attendances/today';

  // HEALTH
  static const health = '/health';
}
