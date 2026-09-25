import React, { useState } from 'react';

interface ProofPortalProps {
  orderId: string;
  previewImageUrl: string;
  dpi: number;
}

export const ProofApprovalPortal: React.FC<ProofPortalProps> = ({ orderId, previewImageUrl, dpi }) => {
  const [status, setStatus] = useState<string>('Pending');
  const [notes, setNotes] = useState<string>('');

  return (
    <div style={{ padding: '24px', fontFamily: 'Arial, sans-serif', maxWidth: '800px', margin: 'auto' }}>
      <h1>Digital Proof Review - Order #{orderId}</h1>
      <p>Quality check: <strong>{dpi} DPI</strong> ({dpi >= 300 ? 'Sharp & Ready' : 'Low Quality Warning'})</p>

      {/* Visual Canvas */}
      <div style={{ position: 'relative', border: '2px solid #ddd', padding: '10px', textAlign: 'center' }}>
        <img src={previewImageUrl} alt="Proof Sample" style={{ maxWidth: '100%', height: 'auto' }} />
      </div>

      <div style={{ marginTop: '20px' }}>
        <textarea
          placeholder="Optional notes or reason for revision..."
          value={notes}
          onChange={(e) => setNotes(e.target.value)}
          rows={3}
          style={{ width: '100%', padding: '8px' }}
        />
      </div>

      <div style={{ display: 'flex', gap: '12px', marginTop: '16px' }}>
        <button
          onClick={() => setStatus('Approved!')}
          style={{ flex: 1, padding: '12px', background: '#2a9d8f', color: 'white', border: 'none', cursor: 'pointer' }}
        >
          Approve Proof
        </button>
        <button
          onClick={() => setStatus('Changes Requested')}
          style={{ flex: 1, padding: '12px', background: '#e76f51', color: 'white', border: 'none', cursor: 'pointer' }}
        >
          Request Changes
        </button>
      </div>

      {status !== 'Pending' && <p style={{ marginTop: '16px', fontWeight: 'bold' }}>Status: {status}</p>}
    </div>
  );
};
