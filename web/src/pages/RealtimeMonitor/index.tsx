import { useState } from 'react';
import { useRealtimeMonitor } from '../../hooks/useRealtimeMonitor';
import type { TrafficData } from '../../hooks/useRealtimeMonitor';
import { MonitorOverview, TrafficChart, ProxyList, ProxyDetailModal } from '../../components/monitor';

export function Component() {
  const {
    trafficData,
    totalInRate,
    totalOutRate,
    onlineCount,
    totalCount,
    clientGroups,
    topProxies,
    chartHistory,
    connected,
    getProxyHistory,
  } = useRealtimeMonitor();

  const [selectedProxy, setSelectedProxy] = useState<TrafficData | null>(null);
  const [modalOpen, setModalOpen] = useState(false);

  const handleProxyClick = (proxy: TrafficData) => {
    setSelectedProxy(proxy);
    setModalOpen(true);
  };

  return (
    <div className="space-y-6 p-6">
      <MonitorOverview
        totalInRate={totalInRate}
        totalOutRate={totalOutRate}
        onlineCount={onlineCount}
        totalCount={totalCount}
        connected={connected}
      />

      <TrafficChart chartHistory={chartHistory} topProxies={topProxies} />

      <ProxyList
        trafficData={trafficData}
        clientGroups={clientGroups}
        getProxyHistory={getProxyHistory}
        onProxyClick={handleProxyClick}
      />

      <ProxyDetailModal
        proxy={selectedProxy}
        history={selectedProxy ? getProxyHistory(selectedProxy.proxy_id) : []}
        open={modalOpen}
        onClose={() => setModalOpen(false)}
      />
    </div>
  );
}