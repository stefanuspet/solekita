import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:dio/dio.dart';
import 'package:mobile/core/network/api_exception.dart';
import 'package:mobile/core/router/app_router.dart';
import 'package:mobile/core/theme/app_colors.dart';
import 'package:mobile/data/local/secure_storage.dart';
import 'package:mobile/data/remote/user_remote.dart';
import 'package:mobile/domain/models/outlet_model.dart';
import 'package:mobile/domain/models/treatment_model.dart';
import 'package:mobile/domain/providers/auth_provider.dart';
import 'package:mobile/domain/providers/outlet_provider.dart';
import 'package:mobile/domain/providers/treatment_provider.dart';

class FirstTimeSetupScreen extends ConsumerStatefulWidget {
  const FirstTimeSetupScreen({super.key});

  @override
  ConsumerState<FirstTimeSetupScreen> createState() =>
      _FirstTimeSetupScreenState();
}

class _FirstTimeSetupScreenState extends ConsumerState<FirstTimeSetupScreen> {
  int _step = 1;
  static const int _totalSteps = 3;

  // ── Step 1 — Outlet ───────────────────────────────────────────────────────
  final _step1FormKey = GlobalKey<FormState>();
  final _outletNameCtrl = TextEditingController();
  final _outletAddressCtrl = TextEditingController();
  bool _outletLoaded = false;
  bool _step1Saving = false;
  String? _step1Error;

  // ── Step 2 — Treatment ────────────────────────────────────────────────────
  final _step2FormKey = GlobalKey<FormState>();
  final _treatmentNameCtrl = TextEditingController();
  final _treatmentPriceCtrl = TextEditingController();
  String? _selectedMaterial;
  final List<TreatmentModel> _treatments = [];
  bool _step2Saving = false;
  String? _step2Error;

  // ── Step 3 — Staff ────────────────────────────────────────────────────────
  final _step3FormKey = GlobalKey<FormState>();
  final _staffNameCtrl = TextEditingController();
  final _staffPhoneCtrl = TextEditingController();
  bool _step3Saving = false;
  String? _step3Error;
  StaffModel? _createdStaff;
  String? _tempPassword;
  bool _passwordCopied = false;

  static const _materials = [
    'Kanvas',
    'Kulit',
    'Suede',
    'Mesh',
    'Rajut',
    'Karet',
    'Lainnya',
  ];

  @override
  void dispose() {
    _outletNameCtrl.dispose();
    _outletAddressCtrl.dispose();
    _treatmentNameCtrl.dispose();
    _treatmentPriceCtrl.dispose();
    _staffNameCtrl.dispose();
    _staffPhoneCtrl.dispose();
    super.dispose();
  }

  void _initOutlet(OutletModel outlet) {
    if (_outletLoaded) return;
    _outletNameCtrl.text = outlet.name;
    _outletAddressCtrl.text = outlet.address;
    _outletLoaded = true;
  }

  // ── Step 1 actions ────────────────────────────────────────────────────────

  Future<void> _submitStep1(OutletModel outlet) async {
    if (!_step1FormKey.currentState!.validate()) return;
    setState(() {
      _step1Saving = true;
      _step1Error = null;
    });
    try {
      final newName = _outletNameCtrl.text.trim();
      final newAddress = _outletAddressCtrl.text.trim();
      final nameChanged = newName != outlet.name;
      final addressChanged = newAddress != outlet.address;
      if (nameChanged || addressChanged) {
        await ref.read(outletRemoteProvider).updateMyOutlet(
              name: nameChanged ? newName : null,
              address: addressChanged ? newAddress : null,
            );
        ref.invalidate(myOutletProvider);
      }
      if (mounted) setState(() => _step = 2);
    } on DioException catch (e) {
      final api = ApiException.from(e);
      if (mounted) setState(() => _step1Error = api?.message ?? 'Terjadi kesalahan');
    } catch (_) {
      if (mounted) setState(() => _step1Error = 'Terjadi kesalahan');
    } finally {
      if (mounted) setState(() => _step1Saving = false);
    }
  }

  // ── Step 2 actions ────────────────────────────────────────────────────────

  Future<void> _addTreatment() async {
    if (!_step2FormKey.currentState!.validate()) return;
    setState(() {
      _step2Saving = true;
      _step2Error = null;
    });
    try {
      final treatment = await ref.read(treatmentRemoteProvider).createTreatment(
            name: _treatmentNameCtrl.text.trim(),
            material: _selectedMaterial!,
            price: int.parse(_treatmentPriceCtrl.text.trim()),
          );
      setState(() {
        _treatments.add(treatment);
        _treatmentNameCtrl.clear();
        _treatmentPriceCtrl.clear();
        _selectedMaterial = null;
      });
      _step2FormKey.currentState?.reset();
    } on DioException catch (e) {
      final api = ApiException.from(e);
      if (mounted) setState(() => _step2Error = api?.message ?? 'Terjadi kesalahan');
    } catch (_) {
      if (mounted) setState(() => _step2Error = 'Terjadi kesalahan');
    } finally {
      if (mounted) setState(() => _step2Saving = false);
    }
  }

  void _nextStep2() {
    if (_treatments.isEmpty) {
      setState(() => _step2Error = 'Tambahkan minimal 1 treatment sebelum lanjut');
      return;
    }
    ref.invalidate(activeTreatmentsProvider);
    setState(() => _step = 3);
  }

  // ── Step 3 actions ────────────────────────────────────────────────────────

  Future<void> _addStaff() async {
    if (!_step3FormKey.currentState!.validate()) return;
    setState(() {
      _step3Saving = true;
      _step3Error = null;
    });
    try {
      final result = await UserRemote(ref.read(apiClientProvider)).createUser(
        name: _staffNameCtrl.text.trim(),
        phone: _staffPhoneCtrl.text.trim(),
        permissions: ['kasir'],
      );
      setState(() {
        _createdStaff = result.user;
        _tempPassword = result.temporaryPassword;
      });
    } on DioException catch (e) {
      final api = ApiException.from(e);
      if (mounted) setState(() => _step3Error = api?.message ?? 'Terjadi kesalahan');
    } catch (_) {
      if (mounted) setState(() => _step3Error = 'Terjadi kesalahan');
    } finally {
      if (mounted) setState(() => _step3Saving = false);
    }
  }

  Future<void> _finish() async {
    await SecureStorage.setSetupCompleted();
    if (!mounted) return;
    context.go(AppRoutes.home);
  }

  // ── Build ─────────────────────────────────────────────────────────────────

  @override
  Widget build(BuildContext context) {
    final textTheme = Theme.of(context).textTheme;
    return Scaffold(
      backgroundColor: AppColors.background,
      body: SafeArea(
        child: Column(
          children: [
            _buildHeader(textTheme),
            Expanded(
              child: SingleChildScrollView(
                padding:
                    const EdgeInsets.symmetric(horizontal: 24, vertical: 24),
                child: _buildStep(textTheme),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildHeader(TextTheme textTheme) {
    return Container(
      color: AppColors.surface,
      padding: const EdgeInsets.fromLTRB(24, 16, 24, 16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            'Langkah $_step dari $_totalSteps',
            style: textTheme.labelMedium?.copyWith(color: AppColors.textHint),
          ),
          const SizedBox(height: 8),
          ClipRRect(
            borderRadius: BorderRadius.circular(4),
            child: LinearProgressIndicator(
              value: _step / _totalSteps,
              minHeight: 6,
              backgroundColor: AppColors.border,
              color: AppColors.primary,
            ),
          ),
          const SizedBox(height: 12),
          Text(
            _stepTitle(),
            style: textTheme.titleMedium?.copyWith(
              fontWeight: FontWeight.bold,
              color: AppColors.textPrimary,
            ),
          ),
          Text(
            _stepSubtitle(),
            style: textTheme.bodySmall?.copyWith(color: AppColors.textHint),
          ),
        ],
      ),
    );
  }

  String _stepTitle() => switch (_step) {
        1 => 'Data Outlet',
        2 => 'Treatment Pertama',
        _ => 'Tambah Karyawan',
      };

  String _stepSubtitle() => switch (_step) {
        1 => 'Verifikasi nama dan alamat bisnis Anda',
        2 => 'Tambahkan minimal 1 jenis treatment',
        _ => 'Opsional — bisa ditambah nanti di pengaturan',
      };

  Widget _buildStep(TextTheme textTheme) => switch (_step) {
        1 => _buildStep1(textTheme),
        2 => _buildStep2(textTheme),
        _ => _buildStep3(textTheme),
      };

  // ── Step 1 ────────────────────────────────────────────────────────────────

  Widget _buildStep1(TextTheme textTheme) {
    final outletAsync = ref.watch(myOutletProvider);
    return outletAsync.when(
      loading: () => const Center(
        child: Padding(
          padding: EdgeInsets.only(top: 80),
          child: CircularProgressIndicator(),
        ),
      ),
      error: (e, _) => Center(
        child: Column(
          children: [
            const SizedBox(height: 40),
            Text('Gagal memuat data outlet',
                style:
                    textTheme.bodyMedium?.copyWith(color: AppColors.error)),
            const SizedBox(height: 12),
            TextButton(
              onPressed: () => ref.invalidate(myOutletProvider),
              child: const Text('Coba lagi'),
            ),
          ],
        ),
      ),
      data: (outlet) {
        _initOutlet(outlet);
        return Form(
          key: _step1FormKey,
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              if (_step1Error != null) ...[
                _errorBanner(_step1Error!, textTheme),
                const SizedBox(height: 16),
              ],
              TextFormField(
                controller: _outletNameCtrl,
                textCapitalization: TextCapitalization.words,
                decoration: const InputDecoration(
                  labelText: 'Nama Bisnis',
                  prefixIcon: Icon(Icons.store_outlined),
                ),
                validator: (v) {
                  if (v == null || v.trim().isEmpty) return 'Nama bisnis wajib diisi';
                  if (v.trim().length < 3) return 'Minimal 3 karakter';
                  return null;
                },
              ),
              const SizedBox(height: 16),
              TextFormField(
                controller: _outletAddressCtrl,
                maxLines: 3,
                textCapitalization: TextCapitalization.sentences,
                decoration: const InputDecoration(
                  labelText: 'Alamat',
                  hintText: 'Jl. Contoh No. 1, Kota',
                  prefixIcon: Icon(Icons.location_on_outlined),
                  alignLabelWithHint: true,
                ),
                validator: (v) {
                  if (v == null || v.trim().isEmpty) return 'Alamat wajib diisi';
                  return null;
                },
              ),
              const SizedBox(height: 28),
              SizedBox(
                height: 52,
                child: ElevatedButton(
                  onPressed: _step1Saving ? null : () => _submitStep1(outlet),
                  child: _step1Saving
                      ? const _LoadingIndicator()
                      : const Text('Lanjut',
                          style: TextStyle(
                              fontSize: 16, fontWeight: FontWeight.w600)),
                ),
              ),
            ],
          ),
        );
      },
    );
  }

  // ── Step 2 ────────────────────────────────────────────────────────────────

  Widget _buildStep2(TextTheme textTheme) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        if (_step2Error != null) ...[
          _errorBanner(_step2Error!, textTheme),
          const SizedBox(height: 16),
        ],
        if (_treatments.isNotEmpty) ...[
          ..._treatments.map(
            (t) => Padding(
              padding: const EdgeInsets.only(bottom: 8),
              child: _TreatmentChip(treatment: t),
            ),
          ),
          const SizedBox(height: 16),
          const Divider(),
          const SizedBox(height: 16),
          Text(
            'Tambah Treatment Lain (opsional)',
            style:
                textTheme.labelMedium?.copyWith(color: AppColors.textHint),
          ),
          const SizedBox(height: 12),
        ],
        Form(
          key: _step2FormKey,
          child: Column(
            children: [
              TextFormField(
                controller: _treatmentNameCtrl,
                textCapitalization: TextCapitalization.words,
                decoration: const InputDecoration(
                  labelText: 'Nama Treatment',
                  hintText: 'Contoh: Cuci Reguler',
                  prefixIcon: Icon(Icons.cleaning_services_outlined),
                ),
                validator: (v) => (v == null || v.trim().isEmpty)
                    ? 'Nama treatment wajib diisi'
                    : null,
              ),
              const SizedBox(height: 16),
              DropdownButtonFormField<String>(
                value: _selectedMaterial,
                decoration: const InputDecoration(
                  labelText: 'Material',
                  prefixIcon: Icon(Icons.category_outlined),
                ),
                items: _materials
                    .map((m) => DropdownMenuItem(value: m, child: Text(m)))
                    .toList(),
                onChanged: (v) => setState(() => _selectedMaterial = v),
                validator: (v) =>
                    v == null ? 'Material wajib dipilih' : null,
              ),
              const SizedBox(height: 16),
              TextFormField(
                controller: _treatmentPriceCtrl,
                keyboardType: TextInputType.number,
                inputFormatters: [FilteringTextInputFormatter.digitsOnly],
                decoration: const InputDecoration(
                  labelText: 'Harga',
                  hintText: '25000',
                  prefixIcon: Icon(Icons.payments_outlined),
                  prefixText: 'Rp ',
                ),
                validator: (v) {
                  if (v == null || v.isEmpty) return 'Harga wajib diisi';
                  final price = int.tryParse(v);
                  if (price == null || price <= 0) return 'Harga tidak valid';
                  return null;
                },
              ),
            ],
          ),
        ),
        const SizedBox(height: 20),
        OutlinedButton(
          onPressed: _step2Saving ? null : _addTreatment,
          child: _step2Saving
              ? const _LoadingIndicator(color: AppColors.primary)
              : Text(
                  _treatments.isEmpty ? 'Tambah Treatment' : 'Tambah Lagi',
                  style: const TextStyle(fontWeight: FontWeight.w600),
                ),
        ),
        const SizedBox(height: 12),
        SizedBox(
          height: 52,
          child: ElevatedButton(
            onPressed: _step2Saving ? null : _nextStep2,
            child: const Text('Lanjut',
                style:
                    TextStyle(fontSize: 16, fontWeight: FontWeight.w600)),
          ),
        ),
      ],
    );
  }

  // ── Step 3 ────────────────────────────────────────────────────────────────

  Widget _buildStep3(TextTheme textTheme) {
    if (_createdStaff != null && _tempPassword != null) {
      return _buildPasswordReveal(textTheme);
    }
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        if (_step3Error != null) ...[
          _errorBanner(_step3Error!, textTheme),
          const SizedBox(height: 16),
        ],
        Form(
          key: _step3FormKey,
          child: Column(
            children: [
              TextFormField(
                controller: _staffNameCtrl,
                textCapitalization: TextCapitalization.words,
                decoration: const InputDecoration(
                  labelText: 'Nama Karyawan',
                  prefixIcon: Icon(Icons.person_outline),
                ),
                validator: (v) => (v == null || v.trim().isEmpty)
                    ? 'Nama wajib diisi'
                    : null,
              ),
              const SizedBox(height: 16),
              TextFormField(
                controller: _staffPhoneCtrl,
                keyboardType: TextInputType.phone,
                decoration: const InputDecoration(
                  labelText: 'Nomor HP Karyawan',
                  hintText: '08xxxxxxxxxx',
                  prefixIcon: Icon(Icons.phone_outlined),
                ),
                validator: (v) {
                  if (v == null || v.trim().isEmpty) {
                    return 'Nomor HP wajib diisi';
                  }
                  if (!RegExp(r'^08\d{8,11}$').hasMatch(v.trim())) {
                    return 'Format nomor HP tidak valid';
                  }
                  return null;
                },
              ),
            ],
          ),
        ),
        const SizedBox(height: 20),
        Container(
          padding: const EdgeInsets.all(12),
          decoration: BoxDecoration(
            color: AppColors.surfaceTeal,
            borderRadius: BorderRadius.circular(12),
          ),
          child: Row(
            children: [
              const Icon(Icons.info_outline,
                  color: AppColors.primary, size: 18),
              const SizedBox(width: 8),
              Expanded(
                child: Text(
                  'Karyawan akan mendapat role Kasir. Password sementara ditampilkan setelah ini.',
                  style: textTheme.bodySmall
                      ?.copyWith(color: AppColors.primaryDark),
                ),
              ),
            ],
          ),
        ),
        const SizedBox(height: 20),
        SizedBox(
          height: 52,
          child: ElevatedButton(
            onPressed: _step3Saving ? null : _addStaff,
            child: _step3Saving
                ? const _LoadingIndicator()
                : const Text('Tambah Karyawan',
                    style: TextStyle(
                        fontSize: 16, fontWeight: FontWeight.w600)),
          ),
        ),
        const SizedBox(height: 12),
        TextButton(
          onPressed: _step3Saving ? null : _finish,
          child: const Text(
            'Lewati, tambah karyawan nanti',
            style: TextStyle(color: AppColors.textHint),
          ),
        ),
      ],
    );
  }

  Widget _buildPasswordReveal(TextTheme textTheme) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Center(
          child: Container(
            width: 64,
            height: 64,
            decoration: const BoxDecoration(
              color: AppColors.surfaceTeal,
              shape: BoxShape.circle,
            ),
            child: const Icon(Icons.check_circle_outline,
                color: AppColors.primary, size: 36),
          ),
        ),
        const SizedBox(height: 16),
        Text(
          '${_createdStaff!.name} berhasil ditambahkan',
          style: textTheme.titleMedium?.copyWith(
            fontWeight: FontWeight.bold,
            color: AppColors.textPrimary,
          ),
          textAlign: TextAlign.center,
        ),
        const SizedBox(height: 8),
        Text(
          'Simpan password sementara ini. Karyawan wajib ganti password setelah login pertama.',
          style: textTheme.bodySmall?.copyWith(color: AppColors.textHint),
          textAlign: TextAlign.center,
        ),
        const SizedBox(height: 24),
        Container(
          padding:
              const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
          decoration: BoxDecoration(
            color: AppColors.surface,
            borderRadius: BorderRadius.circular(12),
            border: Border.all(color: AppColors.border),
          ),
          child: Row(
            children: [
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text('Password Sementara',
                        style: textTheme.labelSmall
                            ?.copyWith(color: AppColors.textHint)),
                    const SizedBox(height: 4),
                    SelectableText(
                      _tempPassword!,
                      style: textTheme.titleMedium?.copyWith(
                        fontFamily: 'monospace',
                        letterSpacing: 2,
                        fontWeight: FontWeight.bold,
                        color: AppColors.textPrimary,
                      ),
                    ),
                  ],
                ),
              ),
              IconButton(
                onPressed: () {
                  Clipboard.setData(ClipboardData(text: _tempPassword!));
                  setState(() => _passwordCopied = true);
                  ScaffoldMessenger.of(context).showSnackBar(
                    const SnackBar(
                      content: Text('Password disalin'),
                      duration: Duration(seconds: 2),
                    ),
                  );
                },
                icon: Icon(
                  _passwordCopied
                      ? Icons.check
                      : Icons.copy_outlined,
                  color: _passwordCopied
                      ? AppColors.success
                      : AppColors.textHint,
                ),
                tooltip: 'Salin password',
              ),
            ],
          ),
        ),
        const SizedBox(height: 32),
        SizedBox(
          height: 52,
          child: ElevatedButton(
            onPressed: _finish,
            child: const Text(
              'Selesai, Mulai Gunakan Solekita',
              style:
                  TextStyle(fontSize: 16, fontWeight: FontWeight.w600),
            ),
          ),
        ),
      ],
    );
  }

  Widget _errorBanner(String message, TextTheme textTheme) {
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: AppColors.error.withOpacity(0.08),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: AppColors.error.withOpacity(0.3)),
      ),
      child: Row(
        children: [
          const Icon(Icons.error_outline, color: AppColors.error, size: 18),
          const SizedBox(width: 8),
          Expanded(
            child: Text(message,
                style:
                    textTheme.bodySmall?.copyWith(color: AppColors.error)),
          ),
        ],
      ),
    );
  }
}

// ── Private helper widgets ─────────────────────────────────────────────────

class _LoadingIndicator extends StatelessWidget {
  final Color color;
  const _LoadingIndicator({this.color = Colors.white});

  @override
  Widget build(BuildContext context) => SizedBox(
        width: 22,
        height: 22,
        child: CircularProgressIndicator(color: color, strokeWidth: 2.5),
      );
}

class _TreatmentChip extends StatelessWidget {
  final TreatmentModel treatment;
  const _TreatmentChip({required this.treatment});

  @override
  Widget build(BuildContext context) {
    final textTheme = Theme.of(context).textTheme;
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
      decoration: BoxDecoration(
        color: AppColors.surfaceTeal,
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: AppColors.primary.withOpacity(0.3)),
      ),
      child: Row(
        children: [
          const Icon(Icons.check_circle, color: AppColors.primary, size: 18),
          const SizedBox(width: 10),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  treatment.name,
                  style: textTheme.bodyMedium?.copyWith(
                      fontWeight: FontWeight.w600,
                      color: AppColors.textPrimary),
                ),
                Text(
                  '${treatment.material} • Rp ${_formatPrice(treatment.price)}',
                  style: textTheme.bodySmall
                      ?.copyWith(color: AppColors.textHint),
                ),
              ],
            ),
          ),
        ],
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
