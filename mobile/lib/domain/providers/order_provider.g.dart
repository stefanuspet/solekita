// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'order_provider.dart';

// **************************************************************************
// RiverpodGenerator
// **************************************************************************

String _$orderRemoteHash() => r'147b76a38a3beeaaca60422bec2a1c665a4b3df1';

/// See also [orderRemote].
@ProviderFor(orderRemote)
final orderRemoteProvider = AutoDisposeProvider<OrderRemote>.internal(
  orderRemote,
  name: r'orderRemoteProvider',
  debugGetCreateSourceHash:
      const bool.fromEnvironment('dart.vm.product') ? null : _$orderRemoteHash,
  dependencies: null,
  allTransitiveDependencies: null,
);

typedef OrderRemoteRef = AutoDisposeProviderRef<OrderRemote>;
String _$todayOrdersHash() => r'add9a38b8323667386cae3657c3dde9b804f368f';

/// Copied from Dart SDK
class _SystemHash {
  _SystemHash._();

  static int combine(int hash, int value) {
    // ignore: parameter_assignments
    hash = 0x1fffffff & (hash + value);
    // ignore: parameter_assignments
    hash = 0x1fffffff & (hash + ((0x0007ffff & hash) << 10));
    return hash ^ (hash >> 6);
  }

  static int finish(int hash) {
    // ignore: parameter_assignments
    hash = 0x1fffffff & (hash + ((0x03ffffff & hash) << 3));
    // ignore: parameter_assignments
    hash = hash ^ (hash >> 11);
    return 0x1fffffff & (hash + ((0x00003fff & hash) << 15));
  }
}

/// List order hari ini. Invalidate untuk refresh: ref.invalidate(todayOrdersProvider)
///
/// Copied from [todayOrders].
@ProviderFor(todayOrders)
const todayOrdersProvider = TodayOrdersFamily();

/// List order hari ini. Invalidate untuk refresh: ref.invalidate(todayOrdersProvider)
///
/// Copied from [todayOrders].
class TodayOrdersFamily extends Family<AsyncValue<OrderListResult>> {
  /// List order hari ini. Invalidate untuk refresh: ref.invalidate(todayOrdersProvider)
  ///
  /// Copied from [todayOrders].
  const TodayOrdersFamily();

  /// List order hari ini. Invalidate untuk refresh: ref.invalidate(todayOrdersProvider)
  ///
  /// Copied from [todayOrders].
  TodayOrdersProvider call({
    String? status,
    String? search,
  }) {
    return TodayOrdersProvider(
      status: status,
      search: search,
    );
  }

  @override
  TodayOrdersProvider getProviderOverride(
    covariant TodayOrdersProvider provider,
  ) {
    return call(
      status: provider.status,
      search: provider.search,
    );
  }

  static const Iterable<ProviderOrFamily>? _dependencies = null;

  @override
  Iterable<ProviderOrFamily>? get dependencies => _dependencies;

  static const Iterable<ProviderOrFamily>? _allTransitiveDependencies = null;

  @override
  Iterable<ProviderOrFamily>? get allTransitiveDependencies =>
      _allTransitiveDependencies;

  @override
  String? get name => r'todayOrdersProvider';
}

/// List order hari ini. Invalidate untuk refresh: ref.invalidate(todayOrdersProvider)
///
/// Copied from [todayOrders].
class TodayOrdersProvider extends AutoDisposeFutureProvider<OrderListResult> {
  /// List order hari ini. Invalidate untuk refresh: ref.invalidate(todayOrdersProvider)
  ///
  /// Copied from [todayOrders].
  TodayOrdersProvider({
    String? status,
    String? search,
  }) : this._internal(
          (ref) => todayOrders(
            ref as TodayOrdersRef,
            status: status,
            search: search,
          ),
          from: todayOrdersProvider,
          name: r'todayOrdersProvider',
          debugGetCreateSourceHash:
              const bool.fromEnvironment('dart.vm.product')
                  ? null
                  : _$todayOrdersHash,
          dependencies: TodayOrdersFamily._dependencies,
          allTransitiveDependencies:
              TodayOrdersFamily._allTransitiveDependencies,
          status: status,
          search: search,
        );

  TodayOrdersProvider._internal(
    super._createNotifier, {
    required super.name,
    required super.dependencies,
    required super.allTransitiveDependencies,
    required super.debugGetCreateSourceHash,
    required super.from,
    required this.status,
    required this.search,
  }) : super.internal();

  final String? status;
  final String? search;

  @override
  Override overrideWith(
    FutureOr<OrderListResult> Function(TodayOrdersRef provider) create,
  ) {
    return ProviderOverride(
      origin: this,
      override: TodayOrdersProvider._internal(
        (ref) => create(ref as TodayOrdersRef),
        from: from,
        name: null,
        dependencies: null,
        allTransitiveDependencies: null,
        debugGetCreateSourceHash: null,
        status: status,
        search: search,
      ),
    );
  }

  @override
  AutoDisposeFutureProviderElement<OrderListResult> createElement() {
    return _TodayOrdersProviderElement(this);
  }

  @override
  bool operator ==(Object other) {
    return other is TodayOrdersProvider &&
        other.status == status &&
        other.search == search;
  }

  @override
  int get hashCode {
    var hash = _SystemHash.combine(0, runtimeType.hashCode);
    hash = _SystemHash.combine(hash, status.hashCode);
    hash = _SystemHash.combine(hash, search.hashCode);

    return _SystemHash.finish(hash);
  }
}

mixin TodayOrdersRef on AutoDisposeFutureProviderRef<OrderListResult> {
  /// The parameter `status` of this provider.
  String? get status;

  /// The parameter `search` of this provider.
  String? get search;
}

class _TodayOrdersProviderElement
    extends AutoDisposeFutureProviderElement<OrderListResult>
    with TodayOrdersRef {
  _TodayOrdersProviderElement(super.provider);

  @override
  String? get status => (origin as TodayOrdersProvider).status;
  @override
  String? get search => (origin as TodayOrdersProvider).search;
}

String _$orderDetailHash() => r'3b79bdbbe57d36ea0dd511efdce6922b8f4f27f4';

/// Detail order by ID. Family provider — satu instance per orderId.
///
/// Copied from [orderDetail].
@ProviderFor(orderDetail)
const orderDetailProvider = OrderDetailFamily();

/// Detail order by ID. Family provider — satu instance per orderId.
///
/// Copied from [orderDetail].
class OrderDetailFamily extends Family<AsyncValue<OrderModel>> {
  /// Detail order by ID. Family provider — satu instance per orderId.
  ///
  /// Copied from [orderDetail].
  const OrderDetailFamily();

  /// Detail order by ID. Family provider — satu instance per orderId.
  ///
  /// Copied from [orderDetail].
  OrderDetailProvider call(
    String orderId,
  ) {
    return OrderDetailProvider(
      orderId,
    );
  }

  @override
  OrderDetailProvider getProviderOverride(
    covariant OrderDetailProvider provider,
  ) {
    return call(
      provider.orderId,
    );
  }

  static const Iterable<ProviderOrFamily>? _dependencies = null;

  @override
  Iterable<ProviderOrFamily>? get dependencies => _dependencies;

  static const Iterable<ProviderOrFamily>? _allTransitiveDependencies = null;

  @override
  Iterable<ProviderOrFamily>? get allTransitiveDependencies =>
      _allTransitiveDependencies;

  @override
  String? get name => r'orderDetailProvider';
}

/// Detail order by ID. Family provider — satu instance per orderId.
///
/// Copied from [orderDetail].
class OrderDetailProvider extends AutoDisposeFutureProvider<OrderModel> {
  /// Detail order by ID. Family provider — satu instance per orderId.
  ///
  /// Copied from [orderDetail].
  OrderDetailProvider(
    String orderId,
  ) : this._internal(
          (ref) => orderDetail(
            ref as OrderDetailRef,
            orderId,
          ),
          from: orderDetailProvider,
          name: r'orderDetailProvider',
          debugGetCreateSourceHash:
              const bool.fromEnvironment('dart.vm.product')
                  ? null
                  : _$orderDetailHash,
          dependencies: OrderDetailFamily._dependencies,
          allTransitiveDependencies:
              OrderDetailFamily._allTransitiveDependencies,
          orderId: orderId,
        );

  OrderDetailProvider._internal(
    super._createNotifier, {
    required super.name,
    required super.dependencies,
    required super.allTransitiveDependencies,
    required super.debugGetCreateSourceHash,
    required super.from,
    required this.orderId,
  }) : super.internal();

  final String orderId;

  @override
  Override overrideWith(
    FutureOr<OrderModel> Function(OrderDetailRef provider) create,
  ) {
    return ProviderOverride(
      origin: this,
      override: OrderDetailProvider._internal(
        (ref) => create(ref as OrderDetailRef),
        from: from,
        name: null,
        dependencies: null,
        allTransitiveDependencies: null,
        debugGetCreateSourceHash: null,
        orderId: orderId,
      ),
    );
  }

  @override
  AutoDisposeFutureProviderElement<OrderModel> createElement() {
    return _OrderDetailProviderElement(this);
  }

  @override
  bool operator ==(Object other) {
    return other is OrderDetailProvider && other.orderId == orderId;
  }

  @override
  int get hashCode {
    var hash = _SystemHash.combine(0, runtimeType.hashCode);
    hash = _SystemHash.combine(hash, orderId.hashCode);

    return _SystemHash.finish(hash);
  }
}

mixin OrderDetailRef on AutoDisposeFutureProviderRef<OrderModel> {
  /// The parameter `orderId` of this provider.
  String get orderId;
}

class _OrderDetailProviderElement
    extends AutoDisposeFutureProviderElement<OrderModel> with OrderDetailRef {
  _OrderDetailProviderElement(super.provider);

  @override
  String get orderId => (origin as OrderDetailProvider).orderId;
}

String _$createOrderNotifierHash() =>
    r'16a6c7345bdb9bfef7236d19fb0dca8706b9a6c7';

/// See also [CreateOrderNotifier].
@ProviderFor(CreateOrderNotifier)
final createOrderNotifierProvider = AutoDisposeNotifierProvider<
    CreateOrderNotifier, AsyncValue<OrderModel?>>.internal(
  CreateOrderNotifier.new,
  name: r'createOrderNotifierProvider',
  debugGetCreateSourceHash: const bool.fromEnvironment('dart.vm.product')
      ? null
      : _$createOrderNotifierHash,
  dependencies: null,
  allTransitiveDependencies: null,
);

typedef _$CreateOrderNotifier = AutoDisposeNotifier<AsyncValue<OrderModel?>>;
// ignore_for_file: type=lint
// ignore_for_file: subtype_of_sealed_class, invalid_use_of_internal_member, invalid_use_of_visible_for_testing_member
