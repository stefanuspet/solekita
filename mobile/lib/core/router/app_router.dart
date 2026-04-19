import 'package:go_router/go_router.dart';
import 'package:mobile/data/local/secure_storage.dart';
import 'package:mobile/presentation/screens/attendance/attendance_screen.dart';
import 'package:mobile/presentation/screens/auth/login_screen.dart';
import 'package:mobile/presentation/screens/auth/register_screen.dart';
import 'package:mobile/presentation/screens/home/home_screen.dart';
import 'package:mobile/presentation/screens/order/create_order_screen.dart';
import 'package:mobile/presentation/screens/order/order_detail_screen.dart';
import 'package:mobile/presentation/screens/settings/settings_screen.dart';
import 'package:mobile/presentation/screens/setup/first_time_setup_screen.dart';
import 'package:mobile/presentation/screens/splash_screen.dart';

class AppRoutes {
  static const splash = '/';
  static const login = '/login';
  static const register = '/register';
  static const firstTimeSetup = '/setup';
  static const home = '/home';
  static const createOrder = '/order/create';
  static const orderDetail = '/order/:id';
  static const attendance = '/attendance';
  static const settings = '/settings';

  // Route yang bisa diakses tanpa token
  static const _publicRoutes = {login, register};

  // Route yang tidak terkena redirect setelah cek setup
  static const _setupExemptRoutes = {splash, login, register, firstTimeSetup};
}

final appRouter = GoRouter(
  initialLocation: AppRoutes.splash,
  redirect: (context, state) async {
    final loc = state.matchedLocation;

    // Splash menangani redirect-nya sendiri
    if (loc == AppRoutes.splash) return null;

    final token = await SecureStorage.getAccessToken();

    // ── Tidak ada token ────────────────────────────────────────────────────
    if (token == null) {
      // Halaman public boleh diakses
      if (AppRoutes._publicRoutes.contains(loc)) return null;
      return AppRoutes.login;
    }

    // ── Ada token ──────────────────────────────────────────────────────────
    // Jangan bisa kembali ke halaman auth
    if (AppRoutes._publicRoutes.contains(loc)) return AppRoutes.home;

    // Cek setup hanya untuk route di luar exempted list
    if (!AppRoutes._setupExemptRoutes.contains(loc)) {
      final setupDone = await SecureStorage.isSetupCompleted();
      if (!setupDone) return AppRoutes.firstTimeSetup;
    }

    return null;
  },
  routes: [
    GoRoute(
      path: AppRoutes.splash,
      builder: (context, state) => const SplashScreen(),
    ),
    GoRoute(
      path: AppRoutes.login,
      builder: (context, state) => const LoginScreen(),
    ),
    GoRoute(
      path: AppRoutes.register,
      builder: (context, state) => const RegisterScreen(),
    ),
    GoRoute(
      path: AppRoutes.firstTimeSetup,
      builder: (context, state) => const FirstTimeSetupScreen(),
    ),
    GoRoute(
      path: AppRoutes.home,
      builder: (context, state) => const HomeScreen(),
    ),
    GoRoute(
      path: AppRoutes.createOrder,
      builder: (context, state) => const CreateOrderScreen(),
    ),
    GoRoute(
      path: AppRoutes.orderDetail,
      builder: (context, state) {
        final id = state.pathParameters['id']!;
        return OrderDetailScreen(orderId: id);
      },
    ),
    GoRoute(
      path: AppRoutes.attendance,
      builder: (context, state) => const AttendanceScreen(),
    ),
    GoRoute(
      path: AppRoutes.settings,
      builder: (context, state) => const SettingsScreen(),
    ),
  ],
);
