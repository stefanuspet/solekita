import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:mobile/core/router/app_router.dart';
import 'package:mobile/core/theme/app_colors.dart';
import 'package:mobile/data/remote/order_remote.dart';
import 'package:mobile/domain/models/order_model.dart';
import 'package:mobile/domain/providers/auth_provider.dart';
import 'package:mobile/domain/providers/order_provider.dart';
import 'package:mobile/presentation/widgets/order/status_badge.dart';

class HomeScreen extends ConsumerStatefulWidget {
  const HomeScreen({super.key});

  @override
  ConsumerState<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends ConsumerState<HomeScreen> {
  int _tabIndex = 0;

  @override
  Widget build(BuildContext context) {
    final userAsync = ref.watch(authNotifierProvider);
    final user = userAsync.valueOrNull;
    final isOwner = user?.isOwner ?? false;

    final tabs = _buildTabs(isOwner);

    return Scaffold(
      backgroundColor: AppColors.background,
      appBar: AppBar(
        backgroundColor: AppColors.surface,
        elevation: 0,
        titleSpacing: 16,
        title: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              user?.outlet.name ?? 'Solekita',
              style: const TextStyle(
                fontSize: 16,
                fontWeight: FontWeight.bold,
                color: AppColors.textPrimary,
              ),
            ),
            Text(
              tabs[_tabIndex].label,
              style: const TextStyle(
                fontSize: 12,
                color: AppColors.textHint,
                fontWeight: FontWeight.normal,
              ),
            ),
          ],
        ),
        actions: [
          Padding(
            padding: const EdgeInsets.only(right: 12),
            child: CircleAvatar(
              radius: 18,
              backgroundColor: AppColors.surfaceTeal,
              child: Text(
                _initials(user?.name ?? '?'),
                style: const TextStyle(
                  color: AppColors.primaryDark,
                  fontWeight: FontWeight.bold,
                  fontSize: 13,
                ),
              ),
            ),
          ),
        ],
      ),
      body: IndexedStack(
        index: _tabIndex,
        children: tabs.map((t) => t.body).toList(),
      ),
      floatingActionButton: _tabIndex == 0
          ? FloatingActionButton.extended(
              onPressed: () => context.push(AppRoutes.createOrder),
              backgroundColor: AppColors.primary,
              icon: const Icon(Icons.add, color: Colors.white),
              label: const Text(
                'Order Baru',
                style: TextStyle(
                  color: Colors.white,
                  fontWeight: FontWeight.w600,
                  fontSize: 15,
                ),
              ),
            )
          : null,
      floatingActionButtonLocation: FloatingActionButtonLocation.centerFloat,
      bottomNavigationBar: BottomNavigationBar(
        currentIndex: _tabIndex,
        onTap: (i) => setState(() => _tabIndex = i),
        selectedItemColor: AppColors.primary,
        unselectedItemColor: AppColors.textHint,
        backgroundColor: AppColors.surface,
        type: BottomNavigationBarType.fixed,
        items: tabs.map((t) => t.navItem).toList(),
      ),
    );
  }

  List<_Tab> _buildTabs(bool isOwner) => [
        _Tab(
          label: 'Order Hari Ini',
          navItem: const BottomNavigationBarItem(
            icon: Icon(Icons.receipt_long_outlined),
            activeIcon: Icon(Icons.receipt_long),
            label: 'Order',
          ),
          body: const _OrderTab(),
        ),
        _Tab(
          label: 'Absensi',
          navItem: const BottomNavigationBarItem(
            icon: Icon(Icons.fingerprint_outlined),
            activeIcon: Icon(Icons.fingerprint),
            label: 'Absensi',
          ),
          body: const _AttendanceTab(),
        ),
        if (isOwner)
          _Tab(
            label: 'Pengaturan',
            navItem: const BottomNavigationBarItem(
              icon: Icon(Icons.settings_outlined),
              activeIcon: Icon(Icons.settings),
              label: 'Pengaturan',
            ),
            body: const _SettingsTab(),
          ),
      ];

  String _initials(String name) {
    final parts = name.trim().split(' ');
    if (parts.isEmpty) return '?';
    if (parts.length == 1) return parts[0][0].toUpperCase();
    return '${parts[0][0]}${parts[1][0]}'.toUpperCase();
  }
}

class _Tab {
  final String label;
  final BottomNavigationBarItem navItem;
  final Widget body;
  const _Tab({required this.label, required this.navItem, required this.body});
}

// ── Order Tab ──────────────────────────────────────────────────────────────

class _OrderTab extends ConsumerWidget {
  const _OrderTab();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final ordersAsync = ref.watch(todayOrdersProvider());

    return RefreshIndicator(
      color: AppColors.primary,
      onRefresh: () => ref.refresh(todayOrdersProvider().future),
      child: ordersAsync.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => _ErrorView(
          message: 'Gagal memuat order',
          onRetry: () => ref.invalidate(todayOrdersProvider),
        ),
        data: (result) {
          if (result.items.isEmpty) {
            return _EmptyOrders(
              onRefresh: () => ref.invalidate(todayOrdersProvider),
            );
          }
          return ListView.separated(
            padding: const EdgeInsets.fromLTRB(16, 16, 16, 100),
            itemCount: result.items.length,
            separatorBuilder: (_, __) => const SizedBox(height: 8),
            itemBuilder: (context, i) => _OrderCard(order: result.items[i]),
          );
        },
      ),
    );
  }
}

class _OrderCard extends StatelessWidget {
  final OrderListItem order;
  const _OrderCard({required this.order});

  @override
  Widget build(BuildContext context) {
    final textTheme = Theme.of(context).textTheme;

    return Card(
      margin: EdgeInsets.zero,
      elevation: 0,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(12),
        side: const BorderSide(color: AppColors.border),
      ),
      color: AppColors.surface,
      child: InkWell(
        borderRadius: BorderRadius.circular(12),
        onTap: () {
          final route = AppRoutes.orderDetail.replaceFirst(':id', order.id);
          context.push(route);
        },
        child: Padding(
          padding: const EdgeInsets.all(14),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  Text(
                    order.orderNumber,
                    style: textTheme.labelMedium?.copyWith(
                      color: AppColors.textHint,
                      fontFamily: 'monospace',
                    ),
                  ),
                  const Spacer(),
                  _StatusChip(status: order.status),
                ],
              ),
              const SizedBox(height: 8),
              Text(
                order.customer.name,
                style: textTheme.bodyMedium?.copyWith(
                  fontWeight: FontWeight.w600,
                  color: AppColors.textPrimary,
                ),
              ),
              const SizedBox(height: 2),
              Text(
                '${order.treatmentName} • ${order.material}',
                style:
                    textTheme.bodySmall?.copyWith(color: AppColors.textHint),
              ),
              const SizedBox(height: 10),
              Row(
                children: [
                  Text(
                    'Rp ${_formatPrice(order.totalPrice)}',
                    style: textTheme.bodyMedium?.copyWith(
                      fontWeight: FontWeight.bold,
                      color: AppColors.textPrimary,
                    ),
                  ),
                  const Spacer(),
                  if (order.isPickup)
                    _BadgeIcon(Icons.directions_car_outlined, 'Jemput'),
                  if (order.isDelivery)
                    _BadgeIcon(Icons.delivery_dining_outlined, 'Antar'),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }

  String _formatPrice(int price) {
    final s = price.toString();
    final buf = StringBuffer();
    for (var i = 0; i < s.length; i++) {
      if (i > 0 && (s.length - i) % 3 == 0) buf.write('.');
      buf.write(s[i]);
    }
    return buf.toString();
  }
}

class _StatusChip extends StatelessWidget {
  final OrderStatus status;
  const _StatusChip({required this.status});

  @override
  Widget build(BuildContext context) {
    final (bg, fg) = _colors(status);
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
      decoration: BoxDecoration(
        color: bg,
        borderRadius: BorderRadius.circular(20),
      ),
      child: Text(
        status.label,
        style: TextStyle(
          color: fg,
          fontSize: 11,
          fontWeight: FontWeight.w600,
        ),
      ),
    );
  }

  (Color bg, Color fg) _colors(OrderStatus s) => switch (s) {
        OrderStatus.dijemput =>
          (const Color(0xFFDBEAFE), const Color(0xFF1D4ED8)),
        OrderStatus.baru => (AppColors.surfaceTeal, AppColors.primaryDark),
        OrderStatus.proses =>
          (const Color(0xFFFEF3C7), const Color(0xFFB45309)),
        OrderStatus.selesai =>
          (const Color(0xFFDCFCE7), const Color(0xFF15803D)),
        OrderStatus.diambil =>
          (AppColors.surfaceAlt, AppColors.textSecondary),
        OrderStatus.diantar =>
          (AppColors.surfaceAlt, AppColors.textSecondary),
        OrderStatus.dibatalkan =>
          (const Color(0xFFFEE2E2), const Color(0xFFB91C1C)),
      };
}

class _BadgeIcon extends StatelessWidget {
  final IconData icon;
  final String label;
  const _BadgeIcon(this.icon, this.label);

  @override
  Widget build(BuildContext context) => Padding(
        padding: const EdgeInsets.only(left: 8),
        child: Row(
          children: [
            Icon(icon, size: 14, color: AppColors.textHint),
            const SizedBox(width: 3),
            Text(label,
                style: const TextStyle(
                    fontSize: 11, color: AppColors.textHint)),
          ],
        ),
      );
}

class _EmptyOrders extends StatelessWidget {
  final VoidCallback onRefresh;
  const _EmptyOrders({required this.onRefresh});

  @override
  Widget build(BuildContext context) {
    final textTheme = Theme.of(context).textTheme;
    return ListView(
      // Wrap in ListView so RefreshIndicator works on empty state
      children: [
        SizedBox(
          height: MediaQuery.of(context).size.height * 0.5,
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              const Icon(Icons.receipt_long_outlined,
                  size: 56, color: AppColors.border),
              const SizedBox(height: 16),
              Text('Belum ada order hari ini',
                  style: textTheme.titleSmall
                      ?.copyWith(color: AppColors.textSecondary)),
              const SizedBox(height: 6),
              Text('Tekan "+ Order Baru" untuk mulai',
                  style: textTheme.bodySmall
                      ?.copyWith(color: AppColors.textHint)),
            ],
          ),
        ),
      ],
    );
  }
}

class _ErrorView extends StatelessWidget {
  final String message;
  final VoidCallback onRetry;
  const _ErrorView({required this.message, required this.onRetry});

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          const Icon(Icons.cloud_off_outlined,
              size: 48, color: AppColors.border),
          const SizedBox(height: 12),
          Text(message,
              style: Theme.of(context)
                  .textTheme
                  .bodyMedium
                  ?.copyWith(color: AppColors.textSecondary)),
          const SizedBox(height: 12),
          TextButton(onPressed: onRetry, child: const Text('Coba lagi')),
        ],
      ),
    );
  }
}

// ── Attendance Tab ─────────────────────────────────────────────────────────

class _AttendanceTab extends StatelessWidget {
  const _AttendanceTab();

  @override
  Widget build(BuildContext context) {
    return const Center(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(Icons.fingerprint, size: 56, color: AppColors.border),
          SizedBox(height: 12),
          Text('Absensi', style: TextStyle(color: AppColors.textSecondary)),
        ],
      ),
    );
  }
}

// ── Settings Tab ───────────────────────────────────────────────────────────

class _SettingsTab extends StatelessWidget {
  const _SettingsTab();

  @override
  Widget build(BuildContext context) {
    return const Center(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(Icons.settings_outlined, size: 56, color: AppColors.border),
          SizedBox(height: 12),
          Text('Pengaturan', style: TextStyle(color: AppColors.textSecondary)),
        ],
      ),
    );
  }
}
