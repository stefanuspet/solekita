import 'package:riverpod_annotation/riverpod_annotation.dart';
import 'package:mobile/data/remote/outlet_remote.dart';
import 'package:mobile/domain/models/outlet_model.dart';
import 'package:mobile/domain/providers/auth_provider.dart';

part 'outlet_provider.g.dart';

@riverpod
OutletRemote outletRemote(OutletRemoteRef ref) =>
    OutletRemote(ref.watch(apiClientProvider));

@riverpod
Future<OutletModel> myOutlet(MyOutletRef ref) =>
    ref.watch(outletRemoteProvider).getMyOutlet();

@riverpod
Future<OutletSummary> outletSummary(OutletSummaryRef ref) =>
    ref.watch(outletRemoteProvider).getSummary();
