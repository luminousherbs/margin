import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import {
  getCollection,
  getCollectionItems,
  deleteCollection,
  removeCollectionItem,
  resolveHandle,
} from "../../api/client";
import { Loader2, ArrowLeft, Trash2, Plus, ExternalLink } from "lucide-react";
import CollectionIcon from "../../components/common/CollectionIcon";
import ShareMenu from "../../components/modals/ShareMenu";
import Card from "../../components/common/Card";
import { useStore } from "@nanostores/react";
import { $user } from "../../store/auth";
import type { Collection, AnnotationItem } from "../../types";
import EditCollectionModal from "../../components/modals/EditCollectionModal";
import { Edit3 } from "lucide-react";

interface CollectionDetailProps {
  handle?: string;
  rkey?: string;
  uri?: string;
  initialCollection?: Collection | null;
  initialItems?: AnnotationItem[];
  resolvedUri?: string;
}

export default function CollectionDetail({
  handle,
  rkey,
  uri,
  initialCollection,
  initialItems,
  resolvedUri,
}: CollectionDetailProps) {
  const user = useStore($user);
  const navigate = useNavigate();
  const [collection, setCollection] = useState<Collection | null>(
    initialCollection || null,
  );
  const [items, setItems] = useState<AnnotationItem[]>(initialItems || []);
  const [loading, setLoading] = useState(!initialCollection);
  const [error, setError] = useState<string | null>(null);
  const [isEditModalOpen, setIsEditModalOpen] = useState(false);

  useEffect(() => {
    if (initialCollection) return;

    const loadData = async () => {
      setLoading(true);
      try {
        let targetUri = resolvedUri || uri;
        if (!targetUri && handle && rkey) {
          if (handle.startsWith("did:")) {
            targetUri = `at://${handle}/at.margin.collection/${rkey}`;
          } else {
            const did = await resolveHandle(handle);
            if (did) {
              targetUri = `at://${did}/at.margin.collection/${rkey}`;
            } else {
              setError("Collection not found");
              setLoading(false);
              return;
            }
          }
        }

        if (targetUri) {
          const col = await getCollection(targetUri);
          if (col) {
            setCollection(col);
            const colItems = await getCollectionItems(col.uri);
            setItems(colItems.filter((i) => i && i.uri));
          } else {
            setError("Collection not found");
          }
        }
      } catch {
        setError("Failed to load collection");
      } finally {
        setLoading(false);
      }
    };

    loadData();
  }, [handle, rkey, uri, initialCollection, resolvedUri]);

  const handleDelete = async () => {
    if (!collection) return;
    if (window.confirm("Delete this collection?")) {
      await deleteCollection(collection.id);
      navigate("/collections");
    }
  };

  const handleRemoveItem = async (item: AnnotationItem) => {
    if (!item.collectionItemUri) return;
    if (!window.confirm("Remove from collection?")) return;
    const success = await removeCollectionItem(item.collectionItemUri);
    if (success) {
      setItems((prev) =>
        prev.filter((i) => i.collectionItemUri !== item.collectionItemUri),
      );
    }
  };

  if (loading) {
    return (
      <div className="flex justify-center py-20">
        <Loader2
          className="animate-spin text-primary-600 dark:text-primary-400"
          size={32}
        />
      </div>
    );
  }

  if (error || !collection) {
    return (
      <div className="text-center py-20 text-red-500 dark:text-red-400">
        {error || "Collection not found"}
      </div>
    );
  }

  const isOwner = user?.did === collection.creator?.did;
  const isSemble = collection.uri?.includes("network.cosmik");

  const sembleUrl = (() => {
    if (!isSemble) return "";
    const parts = collection.uri.split("/");
    const rk = parts[parts.length - 1];
    const h = collection.creator?.handle || "";
    return `https://semble.so/profile/${h}/collections/${rk}`;
  })();

  return (
    <div className="animate-fade-in max-w-2xl mx-auto">
      <a
        href="/collections"
        onClick={(e) => {
          e.preventDefault();
          navigate(-1);
        }}
        className="inline-flex items-center gap-1.5 text-sm font-medium text-surface-500 dark:text-surface-400 hover:text-surface-900 dark:hover:text-white mb-4 transition-colors"
      >
        <ArrowLeft size={16} />
        Collections
      </a>

      <div className="bg-white dark:bg-surface-900 rounded-xl p-4 ring-1 ring-black/5 dark:ring-white/5 mb-4">
        <div className="flex items-start gap-3">
          <div className="p-2 bg-primary-50 dark:bg-primary-900/30 text-primary-600 dark:text-primary-400 rounded-lg">
            <CollectionIcon icon={collection.icon} size={24} />
          </div>
          <div className="flex-1 min-w-0">
            <h1 className="text-xl font-bold text-surface-900 dark:text-white truncate">
              {collection.name}
            </h1>
            {collection.description && (
              <p className="text-surface-600 dark:text-surface-300 text-sm mt-1">
                {collection.description}
              </p>
            )}
            <div className="flex items-center gap-2 mt-2 text-xs text-surface-500 dark:text-surface-400">
              <span className="font-medium bg-surface-100 dark:bg-surface-800 px-2 py-0.5 rounded">
                {items.length} items
              </span>
              <span>
                by{" "}
                <a
                  href={`/profile/${collection.creator?.did}`}
                  className="hover:text-primary-600 dark:hover:text-primary-400 hover:underline transition-colors"
                >
                  {collection.creator?.displayName ||
                    collection.creator?.handle}
                </a>
              </span>
            </div>
          </div>
          <div className="flex items-center gap-1">
            <ShareMenu
              uri={collection.uri}
              handle={collection.creator?.handle}
              type="Collection"
              text={collection.name}
            />
            {isOwner && !isSemble && (
              <>
                <button
                  onClick={() => setIsEditModalOpen(true)}
                  className="p-2 text-surface-400 dark:text-surface-500 hover:text-primary-600 dark:hover:text-primary-400 hover:bg-primary-50 dark:hover:bg-primary-900/20 rounded-lg transition-colors"
                  title="Edit collection"
                >
                  <Edit3 size={18} />
                </button>
                <button
                  onClick={handleDelete}
                  className="p-2 text-surface-400 dark:text-surface-500 hover:text-red-500 dark:hover:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20 rounded-lg transition-colors"
                  title="Delete collection"
                >
                  <Trash2 size={18} />
                </button>
              </>
            )}
            {isSemble && (
              <a
                href={sembleUrl}
                target="_blank"
                rel="noopener noreferrer"
                className="inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium rounded-lg bg-surface-100 dark:bg-surface-800 text-surface-600 dark:text-surface-300 hover:bg-surface-200 dark:hover:bg-surface-700 transition-colors"
              >
                <img src="/semble-logo.svg" alt="" className="w-3.5 h-3.5" />
                View in Semble
                <ExternalLink size={12} />
              </a>
            )}
          </div>
        </div>
      </div>

      <EditCollectionModal
        isOpen={isEditModalOpen}
        onClose={() => setIsEditModalOpen(false)}
        collection={collection}
        onUpdate={(updated) =>
          setCollection({
            ...updated,
            creator: updated.creator || collection.creator,
          })
        }
      />

      <div className="space-y-2">
        {items.length === 0 ? (
          <div className="text-center py-12 text-surface-500 dark:text-surface-400 bg-surface-50 dark:bg-surface-800/50 rounded-xl border border-dashed border-surface-200 dark:border-surface-700">
            <Plus
              size={28}
              className="mx-auto mb-2 text-surface-300 dark:text-surface-600"
            />
            <p className="text-sm">Collection is empty</p>
          </div>
        ) : (
          items.map((item) => (
            <div key={item.uri} className="relative group">
              <Card item={item} hideShare />
              {isOwner && !isSemble && item.collectionItemUri && (
                <button
                  className="absolute top-3 right-3 p-1.5 bg-white/90 dark:bg-surface-800/90 backdrop-blur text-surface-400 dark:text-surface-500 hover:text-red-500 dark:hover:text-red-400 rounded-lg shadow-sm transition-all"
                  onClick={() => handleRemoveItem(item)}
                  title="Remove from collection"
                >
                  <Trash2 size={16} />
                </button>
              )}
            </div>
          ))
        )}
      </div>
    </div>
  );
}
