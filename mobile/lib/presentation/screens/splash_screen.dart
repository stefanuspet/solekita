import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:mobile/core/router/app_router.dart';
import 'package:mobile/core/theme/app_colors.dart';
import 'package:mobile/data/local/secure_storage.dart';

class SplashScreen extends StatefulWidget {
  const SplashScreen({super.key});

  @override
  State<SplashScreen> createState() => _SplashScreenState();
}

class _SplashScreenState extends State<SplashScreen> {
  @override
  void initState() {
    super.initState();
    _redirect();
  }

  Future<void> _redirect() async {
    // Beri waktu sebentar agar splash terlihat
    await Future.delayed(const Duration(milliseconds: 400));
    if (!mounted) return;

    final token = await SecureStorage.getAccessToken();
    if (!mounted) return;

    if (token == null) {
      context.go(AppRoutes.login);
      return;
    }

    final setupDone = await SecureStorage.isSetupCompleted();
    if (!mounted) return;

    context.go(setupDone ? AppRoutes.home : AppRoutes.firstTimeSetup);
  }

  @override
  Widget build(BuildContext context) {
    return const Scaffold(
      backgroundColor: AppColors.primary,
      body: Center(
        child: CircularProgressIndicator(color: Colors.white),
      ),
    );
  }
}
